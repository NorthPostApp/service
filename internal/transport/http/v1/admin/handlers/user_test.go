package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/middleware"
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

func setupUserRouter(handler *UserHandler, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/admin/signin", func(c *gin.Context) {
		if uid != "" {
			c.Set(middleware.UidKey, uid)
		}
		c.Next()
	}, handler.SignInAdminUser)
	return r
}

func TestSignInAdminUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		uid            string // set to context
		mockOutput     *services.SignInAdminUserByIdOutput
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name: "success",
			uid:  "user-123",
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
			name:           "service error",
			uid:            "user-123",
			mockOutput:     (*services.SignInAdminUserByIdOutput)(nil),
			mockError:      errors.New("service failed"),
			expectedStatus: http.StatusInternalServerError,
			expectCall:     true,
		},
		{
			name:           "missing uid",
			uid:            "",
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectCall:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockUserService)
			handler := NewUserHandler(mockSvc, slog.Default())
			router := setupUserRouter(handler, tt.uid)
			if tt.expectCall {
				input := services.SignInAdminUserByIdInput{Uid: tt.uid}
				mockSvc.On("SignInAdminUserById", mock.Anything, input).
					Return(tt.mockOutput, tt.mockError).Once()
			}
			req, _ := http.NewRequest("POST", "/admin/signin", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]json.RawMessage
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "data")
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
