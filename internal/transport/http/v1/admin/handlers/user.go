package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"

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

// SignInAdminUser godoc
// @Summary Sign in admin user
// @Description Sign in an admin user with the idToken and update last login timestamp
// @Tags Admin User
// @Param Authorization header string true "Bearer idToken"
// @Produce json
// @Success 200 {object} dto.SignInAdminUserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/user/signin [post]
func (h *UserHandler) SignInAdminUser(c *gin.Context) {
	uid := c.GetString(middleware.UidKey) // from the middleware
	if uid == "" {
		h.logger.Error(
			"missing user id from the middleware context",
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized id token"})
		return
	}
	input := services.SignInAdminUserByIdInput{
		Uid: uid,
	}
	output, err := h.service.SignInAdminUserById(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to sign in admin user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign in user"})
		return
	}
	response := dto.SignInAdminUserResponse{
		Data: dto.ToAdminUserDTO(output.UserData),
	}
	c.JSON(http.StatusOK, response)
}
