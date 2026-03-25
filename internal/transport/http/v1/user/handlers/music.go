package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type musicService interface {
	GetAllMusicList(ctx context.Context) (*services.GetAllMusicListOutput, error)
	GetPresignedMusicURL(
		ctx context.Context,
		input services.GetPresignedMusicURLInput,
	) (*services.GetPresignedMusicURLOutput, error)
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
// @Tags User Music
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetMusicListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/music/list [get]
func (h *MusicHandler) GetMusicList(c *gin.Context) {
	output, err := h.service.GetAllMusicList(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get all music list", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := dto.GetMusicListResponse{Data: dto.ToMusicDTOs(output.Data)}
	c.JSON(http.StatusOK, response)
}

// GetPresignedMusicURL godoc
// @Summary Get presigned music URL
// @Description Get a presigned URL for a specific music track by genre and track name.
// @Tags User Music
// @Accept json
// @Produce json
// @Param genre path string true "Music genre"
// @Param track path string true "Music track filename"
// @Success 200 {object} dto.GetPresignedMusicURLResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/music/{genre}/{track} [get]
func (h *MusicHandler) GetPresignedMusicURL(c *gin.Context) {
	genre := strings.TrimSpace(c.Param("genre"))
	track := strings.TrimSpace(c.Param("track"))
	validParams := utils.ValidateMusicFilename(c, genre, track, h.logger)
	if !validParams {
		return
	}
	input := services.GetPresignedMusicURLInput{
		Filename: fmt.Sprintf("%s/%s", genre, track),
	}
	output, err := h.service.GetPresignedMusicURL(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("failed to get presigned music url", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get presigned url"})
		return
	}
	response := dto.GetPresignedMusicURLResponse{Data: output.URL}
	c.JSON(http.StatusOK, response)
}
