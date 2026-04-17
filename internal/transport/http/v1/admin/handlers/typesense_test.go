package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/infra"
	"north-post/service/internal/transport/http/v1/dto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTypesenseClient struct {
	mock.Mock
}

func (m *mockTypesenseClient) GetSystemInfo(ctx context.Context) (*infra.TypesenseSystemInfo, error) {
	args := m.Called(ctx)
	var systemInfo *infra.TypesenseSystemInfo
	if value := args.Get(0); value != nil {
		systemInfo, _ = value.(*infra.TypesenseSystemInfo)
	}
	return systemInfo, args.Error(1)
}

func setupTypesenseRouter() (*mockTypesenseClient, *gin.Engine) {
	mockClient := new(mockTypesenseClient)
	handler := NewTypesenseHandler(mockClient, slog.Default())
	router := gin.Default()
	router.GET("/admin/typesense/info", handler.GetSystemInfo)
	return mockClient, router
}

func TestTypesenseHandler_GetSystemInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		url            string
		mockOutput     *infra.TypesenseSystemInfo
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			url:  "/admin/typesense/info",
			mockOutput: &infra.TypesenseSystemInfo{
				Health:                    true,
				SystemCPUActivePercentage: float32(0.5),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed",
			url:            "/admin/typesense/info",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		mockClient, router := setupTypesenseRouter()
		mockClient.On("GetSystemInfo", mock.Anything).Return(tt.mockOutput, tt.mockError).Once()
		req, _ := http.NewRequest("GET", tt.url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, tt.expectedStatus, w.Code)
		if tt.mockError == nil {
			var response dto.GetSystemInfoResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.mockOutput.Health, response.Data.Health)
			assert.Equal(
				t,
				tt.mockOutput.SystemCPUActivePercentage,
				response.Data.SystemCPUActivePercentage)
			mockClient.AssertExpectations(t)
		}
	}
}
