package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"north-post/service/internal/infra"
	"north-post/service/internal/transport/http/v1/dto"

	"github.com/gin-gonic/gin"
)

type typesenseClient interface {
	GetSystemInfo(ctx context.Context) (*infra.TypesenseSystemInfo, error)
}

type TypesenseHandler struct {
	client typesenseClient
	logger *slog.Logger
}

func NewTypesenseHandler(client typesenseClient, logger *slog.Logger) *TypesenseHandler {
	return &TypesenseHandler{
		client: client,
		logger: logger,
	}
}

func (h *TypesenseHandler) GetSystemInfo(c *gin.Context) {
	systemInfo, err := h.client.GetSystemInfo(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get typesense system info", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	systemInfoDTO := dto.ToSystemInfoDTO(systemInfo)
	c.JSON(http.StatusOK, dto.GetSystemInfoResponse{Data: systemInfoDTO})
}
