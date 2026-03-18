package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type musicService interface {
	RefreshMusicList(ctx context.Context) (*services.RefreshMusicListOutput, error)
}

type MusicHandler struct {
	service musicService
	logger  *slog.Logger
}

func NewMusicHandler(service musicService, logger *slog.Logger) *MusicHandler {
	return &MusicHandler{
		service: service,
		logger:  logger,
	}
}

// GetMusicList godoc
// @Summary Get music list
// @Description Get the list of music tracks. Pass refresh=true to force a refresh from the upstream source.
// @Tags Admin Music
// @Accept json
// @Produce json
// @Param refresh query bool false "Whether to refresh the music list"
// @Success 200 {object} dto.GetMusicListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/music/ [get]
func (h *MusicHandler) GetMusicList(c *gin.Context) {
	refreshStr := strings.TrimSpace(c.Query("refresh"))
	shouldRefresh, err := strconv.ParseBool(refreshStr)
	if err != nil && refreshStr != "" {
		h.logger.Error("failed to parse refresh query to boolean", "query", refreshStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refresh parameter"})
		return
	}
	if shouldRefresh {
		output, err := h.service.RefreshMusicList(c.Request.Context())
		if err != nil {
			h.logger.Error("failed to refresh music list", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response := dto.GetMusicListResponse{Data: dto.ToMusicDTOs(output.Data)}
		c.JSON(http.StatusOK, response)
		return
	}
	c.JSON(http.StatusOK, "placeholder")
}
