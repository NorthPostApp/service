package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/middleware"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAddressRequestRepository struct {
	mock.Mock
}

func (m *MockAddressRequestRepository) GetRequestsByStatus(
	ctx context.Context,
	opts *repository.GetRequestsByStatusOptions,
) (*repository.GetRequestsByStatusResponse, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*repository.GetRequestsByStatusResponse), args.Error(1)
}

func setupAddressRequestRouter(handler *AddressRequestHandler, language string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	group := r.Group("/admin/address-request",
		middleware.MockLanguageMiddleware(language),
	)
	group.GET("", handler.GetRequestsByStatus)
	return r
}

func TestGetRequestsByStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		language             string
		status               string
		expectedCall         bool
		mockGetRequestOutput *repository.GetRequestsByStatusResponse
		mockGetRequestError  error
		expectedStatus       int
	}{
		{
			name:         "success",
			language:     "zh",
			status:       "pending",
			expectedCall: true,
			mockGetRequestOutput: &repository.GetRequestsByStatusResponse{
				Requests: []models.AddressRequest{
					{
						ID: "test_id",
					},
				},
				InvalidIDs: []string{},
			},
			mockGetRequestError: nil,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                 "invalid status",
			language:             "zh",
			status:               "invalid",
			expectedCall:         false,
			mockGetRequestOutput: nil,
			mockGetRequestError:  nil,
			expectedStatus:       http.StatusBadRequest,
		},
		{
			name:                 "failed call",
			language:             "zh",
			status:               "completed",
			expectedCall:         true,
			mockGetRequestOutput: nil,
			mockGetRequestError:  errors.New("error"),
			expectedStatus:       http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAddressRequestRepo := new(MockAddressRequestRepository)
			handler := NewAddressRequestHandler(
				mockAddressRequestRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)),
			)
			router := setupAddressRequestRouter(handler, tt.language)
			if tt.expectedCall {
				mockAddressRequestRepo.On("GetRequestsByStatus",
					mock.Anything, mock.Anything).
					Return(tt.mockGetRequestOutput, tt.mockGetRequestError).Once()
			}
			req, _ := http.NewRequest("GET",
				fmt.Sprintf("/admin/address-request?language=%s&status=%s", tt.language, tt.status),
				nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCall == true && tt.mockGetRequestError != nil {
				assert.Contains(t, w.Body.String(), "error")
			} else if tt.expectedCall == true {
				assert.Contains(t, w.Body.String(), tt.mockGetRequestOutput.Requests[0].ID)
			}
			mockAddressRequestRepo.AssertExpectations(t)
		})
	}
}
