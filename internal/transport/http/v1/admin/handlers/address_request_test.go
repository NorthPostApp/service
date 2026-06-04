package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"north-post/service/internal/transport/http/v1/dto"
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

func (m *MockAddressRequestRepository) UpdateRequestData(
	ctx context.Context, opts *repository.UpdateRequestDataOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func setupAddressRequestRouter(handler *AddressRequestHandler, language string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	group := r.Group("/admin/address-request",
		middleware.MockLanguageMiddleware(language),
	)
	group.GET("", handler.GetRequestsByStatus)
	group.POST("update", handler.UpdateRequest)
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

func TestUpdateRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		id             string
		language       string
		expectedCall   bool
		updatedRequest models.AddressRequest
		expectedError  error
		status         int
	}{
		{
			name:           "success",
			id:             "mockID",
			language:       "zh",
			expectedCall:   true,
			updatedRequest: models.AddressRequest{ID: "mockID", Status: "processing"},
			expectedError:  nil,
			status:         http.StatusOK,
		},
		{
			name:           "invalid status",
			id:             "mockID",
			language:       "zh",
			expectedCall:   false,
			updatedRequest: models.AddressRequest{ID: "mockID", Status: "mock"},
			expectedError:  nil,
			status:         http.StatusBadRequest,
		},
		{
			name:           "failed response",
			id:             "mockID",
			language:       "zh",
			expectedCall:   true,
			updatedRequest: models.AddressRequest{ID: "mockID", Status: "completed"},
			expectedError:  errors.New("failed"),
			status:         http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequestRepo := new(MockAddressRequestRepository)
			handler := NewAddressRequestHandler(
				mockRequestRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
			router := setupAddressRequestRouter(handler, tt.language)
			if tt.expectedCall {
				mockRequestRepo.On("UpdateRequestData", mock.Anything, mock.Anything).
					Return(tt.expectedError).Once()
			}
			body, err := json.Marshal(dto.UpdateRequest{
				Language:       models.Language(tt.language),
				ID:             tt.id,
				UpdatedRequest: tt.updatedRequest,
			})
			assert.NoError(t, err)
			req, _ := http.NewRequest("POST", "/admin/address-request/update", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if tt.expectedCall && tt.expectedError != nil {
				assert.Contains(t, w.Body.String(), "error")
			}
			mockRequestRepo.AssertExpectations(t)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}
