package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"
	"north-post/service/internal/transport/http/v1/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AddressBookHandler struct {
	userRepo           userRepository
	addressRepo        addressRepository
	addressRequestRepo addressRequestRepository
	logger             *slog.Logger
}

func NewAddressBookHandler(
	userRepo userRepository,
	addressRepo addressRepository,
	addressRequestRepo addressRequestRepository,
	logger *slog.Logger) *AddressBookHandler {
	return &AddressBookHandler{
		userRepo:           userRepo,
		addressRepo:        addressRepo,
		addressRequestRepo: addressRequestRepo,
		logger:             logger,
	}
}

// UpdateSavedAddresses godoc
// @Summary Update user saved addresses
// @Description Add or remove a saved address for the authenticated user
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param request body dto.UpdateUserSavedAddressesRequest true "Address ID and action (add/remove)"
// @Produce json
// @Success 200 {object} dto.UpdateUserSavedAddressesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book [patch]
func (h *AddressBookHandler) UpdateSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	var req dto.UpdateUserSavedAddressesRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	opts := &repository.UpdateUserSavedAddressesOptions{
		UserID:     uid,
		Language:   language,
		AddressIDs: req.AddressIDs,
		Action:     h.convertUpdateMethod(req.Action),
	}
	output, err := h.userRepo.UpdateUserSavedAddresses(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to update user saved address", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update user saved addresses"})
		return
	}
	response := dto.UpdateUserSavedAddressesResponse{Data: output}
	c.JSON(http.StatusOK, response)
}

// GetSavedAddresses godoc
// @Summary Get user saved addresses
// @Description Retrieve all saved addresses for the authenticated user
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param language query string true "Language code (e.g., en, zh)"
// @Produce json
// @Success 200 {object} dto.GetSavedAddressesResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book [get]
func (h *AddressBookHandler) GetSavedAddresses(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	getSavedAddressesOpts := &repository.GetUserSavedAddressesOptions{Uid: uid, Language: language}
	addressIDs, err := h.userRepo.GetUserSavedAddresses(c.Request.Context(), getSavedAddressesOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	getAddressesOpts := &repository.GetAddressesByIDsOptions{
		Language: language,
		IDs:      addressIDs,
	}
	results, err := h.addressRepo.GetAddressesByIDs(
		c.Request.Context(),
		getAddressesOpts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	// if invalid ids is not empty, remove the invalid items in the background process
	if len(results.InvalidIDs) > 0 {
		h.removeInvalidIDsInBackground(uid, language, results.InvalidIDs)
	}
	response := dto.GetSavedAddressesResponse{Data: dto.ToAddressDTOs(results.Addresses)}
	c.JSON(http.StatusOK, response)
}

// CreateNewRequest godoc
// @Summary Create a new address request
// @Description Create a new address request and add it to the user's address book requests
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param request body dto.CreateNewRequest true "New address request payload"
// @Produce json
// @Success 200 {object} dto.CreateNewRequestResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book/request [post]
func (h *AddressBookHandler) CreateNewRequest(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	var req dto.CreateNewRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}

	// check how many active request does this user have

	// create new request
	createRequestOpts := &repository.CreateRequestOptions{
		Language: language,
		UID:      uid,
		Content:  req.Content,
	}
	newRequestID, err := h.addressRequestRepo.CreateNewRequest(c.Request.Context(), createRequestOpts)
	if err != nil {
		h.logger.Error("failed to create address request",
			"path", "user/handlers/address_book/CreateNewRequest",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	// add the id to user's address book request
	updateUserRequestsOpts := &repository.UpdateUserAddressRequestsOptions{
		Language:   language,
		UserID:     uid,
		RequestIDs: []string{newRequestID},
		Action:     repository.Add,
	}
	_, err = h.userRepo.UpdateUserAddressRequests(c.Request.Context(), updateUserRequestsOpts)
	if err != nil {
		h.logger.Error("failed to add address request to user's address book",
			"path", "user/handlers/address_book/CreateNewRequest",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.CreateNewRequestResponse{Data: newRequestID})
}

// GetRequestsByIDs godoc
// @Summary Get user address requests
// @Description Retrieve all address requests for the authenticated user
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Param language query string true "Language code (e.g., en, zh)"
// @Produce json
// @Success 200 {object} dto.GetRequestByIDsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/address-book/request [get]
func (h *AddressBookHandler) GetRequestsByIDs(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	language := models.Language(c.GetString(middleware.LanguageKey))
	if !validateUser(c, uid, h.logger) {
		return
	}
	logger := h.logger.With(
		"path", "v1.user.handlers.address_book.GetRequestsByIDs",
		"uid", uid,
		"language", language,
	)
	// step 1, get IDs from user repository
	getRequestIDsOpts := &repository.GetUserRequestsOptions{
		Language: language,
		Uid:      uid,
	}
	requests, err := h.userRepo.GetUserRequests(c.Request.Context(), getRequestIDsOpts)
	if err != nil {
		logger.Error("failed to get user request ids", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	// step 2, get request with the ids from the previous step
	getRequestsByIDsOpts := &repository.GetRequestsByIDsOptions{
		Language: language, UID: uid, RequestIDs: requests.IDs}
	requestData, err := h.addressRequestRepo.GetRequestsByIDs(c.Request.Context(), getRequestsByIDsOpts)
	if err != nil {
		logger.Error("failed to get request by ids", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	// step 3, update active request count and remove invalid ids in the background
	if len(requestData.InvalidIDs) != 0 {
		h.asyncRemoveInvalidRequestData(uid, language, requestData.InvalidIDs, logger)
	}
	if requests.ActiveRequestCount != requestData.ActiveRequestsCount {
		h.asyncUpdateUserActiveRequestCount(uid, language, requestData.ActiveRequestsCount, logger)
	}
	c.JSON(http.StatusOK, dto.GetRequestByIDsResponse{Data: requestData.Requests})
}

// ---------- Helper methods ----------
func (h *AddressBookHandler) convertUpdateMethod(action string) repository.UpdateAddressBookAction {
	action = strings.ToLower(action)
	switch action {
	case "add":
		return repository.Add
	case "delete", "remove":
		return repository.Delete
	default:
		return -1
	}
}

func (h *AddressBookHandler) removeInvalidIDsInBackground(uid string, language models.Language, invalidIDs []string) {
	go func(invalidIDs []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := h.userRepo.UpdateUserSavedAddresses(
			ctx,
			&repository.UpdateUserSavedAddressesOptions{
				Language:   language,
				UserID:     uid,
				AddressIDs: invalidIDs,
				Action:     repository.Delete,
			},
		)
		if err != nil {
			h.logger.Error("Failed to remove invalid saved addresses",
				"error", err,
				"invalidIDs", invalidIDs,
			)
		}
	}(invalidIDs)
}

func (h *AddressBookHandler) asyncRemoveInvalidRequestData(
	uid string, language models.Language, invalidIDs []string, logger *slog.Logger) {
	go func(invalidIDs []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := h.userRepo.UpdateUserAddressRequests(
			ctx,
			&repository.UpdateUserAddressRequestsOptions{
				Language:   language,
				UserID:     uid,
				RequestIDs: invalidIDs,
				Action:     repository.Delete,
			},
		)
		if err != nil {
			logger.Error("failed to remove invalid ids", "error", err)
		}
	}(invalidIDs)
}

func (h *AddressBookHandler) asyncUpdateUserActiveRequestCount(
	uid string, language models.Language, count int64, logger *slog.Logger) {
	go func(count int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := h.userRepo.UpdateUserActiveRequestCount(
			ctx,
			&repository.UpdateUserActiveRequestCountOptions{
				Language: language,
				UID:      uid,
				Count:    count,
			},
		)
		if err != nil {
			logger.Error("failed to update user active request count", "error", err)
		}
	}(count)
}
