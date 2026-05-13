package handlers

import (
	"log/slog"
	"net/http"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	repo   userRepository
	logger *slog.Logger
}

func NewUserHandler(repository userRepository, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		repo:   repository,
		logger: logger,
	}
}

// AuthenticateAppUser godoc
// @Summary Authenticate app user
// @Description Authenticate an app user with the idToken and update last login timestamp
// @Tags App User
// @Param Authorization header string true "Bearer idToken"
// @Produce json
// @Success 200 {object} dto.AuthenticateAppUserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/signin [post]
func (h *UserHandler) AuthenticateAppUser(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	if !validateUser(c, uid, h.logger) {
		return
	}
	opts := repository.GetUserByIdOptions{Uid: uid}
	output, err := h.repo.AuthenticateAppUserById(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to authenticate app user", "error", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to authenticate app user"})
		return
	}
	response := dto.AuthenticateAppUserResponse{
		Data: dto.ToAppUserDTO(output),
	}
	c.JSON(http.StatusOK, response)
}
