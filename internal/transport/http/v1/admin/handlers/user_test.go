package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService implements the methods used by UserHandler for testing.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignInAdminUserById(ctx context.Context, input services.SignInAdminUserByIdInput) (*services.SignInAdminUserByIdOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.SignInAdminUserByIdOutput), args.Error(1)
}

func setupUserRouter(handler *UserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/admin/signin", handler.SignInAdminUserById)
	return r
}

func TestSignInAdminUserById(t *testing.T) {
	t.Parallel()
	mockSvc := new(MockUserService)
	handler := NewUserHandler(mockSvc, slog.Default())
	router := setupUserRouter(handler)
	tests := []struct {
		name           string
		body           dto.SignInAdminUserByIdRequest
		rawBody        []byte
		mockOutput     *services.SignInAdminUserByIdOutput
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name: "success",
			body: dto.SignInAdminUserByIdRequest{
				Uid: "uid-1",
			},
			mockOutput: &services.SignInAdminUserByIdOutput{
				UserData: models.AdminUser{
					Email:       "test@example.com",
					DisplayName: "Test User",
					LastLogin:   123,
					ImageUrl:    "https://example.com/avatar.png",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectCall:     true,
		},
		{
			name: "service error",
			body: dto.SignInAdminUserByIdRequest{
				Uid: "uid-1",
			},
			mockOutput:     (*services.SignInAdminUserByIdOutput)(nil),
			mockError:      errors.New("service failed"),
			expectedStatus: http.StatusInternalServerError,
			expectCall:     true,
		},
		{
			name:           "invalid json",
			rawBody:        []byte("invalid"),
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
		{
			name:           "missing uid",
			body:           dto.SignInAdminUserByIdRequest{},
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectCall {
				input := services.SignInAdminUserByIdInput{Uid: tt.body.Uid}
				mockSvc.On("SignInAdminUserById", mock.Anything, input).
					Return(tt.mockOutput, tt.mockError).Once()
			}
			body := tt.rawBody
			if body == nil {
				body, _ = json.Marshal(tt.body)
			}
			req, _ := http.NewRequest("POST", "/admin/signin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectCall {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}
