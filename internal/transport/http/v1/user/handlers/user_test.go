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

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) AuthenticateAppUserById(
	ctx context.Context,
	input services.AuthenticateAppUserByIdInput,
) (*services.AuthenticateAppUserByIdOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthenticateAppUserByIdOutput), args.Error(1)
}

func setupUserRouter(handler *UserHandler, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/user/signin", func(c *gin.Context) {
		if uid != "" {
			c.Set(middleware.UidKey, uid)
		}
		c.Next()
	}, handler.AuthenticateAppUser)
	return r
}

func TestAuthenticateAppUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		uid            string
		mockOutput     *services.AuthenticateAppUserByIdOutput
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name: "success",
			uid:  "user-123",
			mockOutput: &services.AuthenticateAppUserByIdOutput{
				UserData: models.AppUser{
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
			name:           "server error",
			uid:            "user-123",
			mockOutput:     nil,
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
			mockSrv := new(mockUserService)
			handler := NewUserHandler(mockSrv, slog.Default())
			router := setupUserRouter(handler, tt.uid)
			if tt.expectCall {
				input := services.AuthenticateAppUserByIdInput{Uid: tt.uid}
				mockSrv.On("AuthenticateAppUserById", mock.Anything, input).
					Return(tt.mockOutput, tt.mockError).Once()
			}
			req, _ := http.NewRequest("POST", "/user/signin", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]json.RawMessage
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "data")
			}
			mockSrv.AssertExpectations(t)
		})
	}
}
