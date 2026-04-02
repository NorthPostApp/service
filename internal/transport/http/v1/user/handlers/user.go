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
	AuthenticateAppUserById(
		ctx context.Context,
		input services.AuthenticateAppUserByIdInput,
	) (*services.AuthenticateAppUserByIdOutput, error)
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

// AuthenticateAppUser godoc
// @Summary Authenticate app user
// @Description Authenticate an app user with the idToken and update last login timestamp
// @Tags User
// @Param Authorization header string true "Bearer idToken"
// @Produce json
// @Success 200 {object} dto.AuthenticateAppUserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/signin [post]
func (h *UserHandler) AuthenticateAppUser(c *gin.Context) {
	uid := c.GetString(middleware.UidKey)
	if uid == "" {
		h.logger.Error(
			"missing user id from the middleware context",
			"path", c.Request.URL.Path,
			"client_ip", c.ClientIP(),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized id token"})
		return
	}
	input := services.AuthenticateAppUserByIdInput{Uid: uid}
	output, err := h.service.AuthenticateAppUserById(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to authenticate app user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	response := dto.AuthenticateAppUserResponse{
		Data: dto.ToAppUserDTO(output.UserData),
	}
	c.JSON(http.StatusOK, response)
}
