package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/utils"

	"github.com/gin-gonic/gin"
)

type userService interface {
	SignInAdminUserById(ctx context.Context, input services.SignInAdminUserByIdInput) (*services.SignInAdminUserByIdOutput, error)
}

type UserHandler struct {
	service userService
	logger  *slog.Logger
}

func NewUserHandler(service userService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// SignInAdminUserById godoc
// @Summary Sign in admin user by ID
// @Description Sign in an admin user using their UID and update last login timestamp
// @Tags Admin User
// @Accept json
// @Produce json
// @Param request body dto.SignInAdminUserByIdRequest true "Request body"
// @Success 200 {object} dto.SignInAdminUserByIdResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/user/signin [post]
func (h *UserHandler) SignInAdminUserById(c *gin.Context) {
	var req dto.SignInAdminUserByIdRequest
	if !utils.BindJSON(c, &req, h.logger) {
		return
	}
	input := services.SignInAdminUserByIdInput{
		Uid: req.Uid,
	}
	output, err := h.service.SignInAdminUserById(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to sign in admin user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign in user"})
		return
	}
	response := dto.SignInAdminUserByIdResponse{
		Data: dto.ToAdminUserDTO(output.UserData),
	}
	c.JSON(http.StatusOK, response)
}
