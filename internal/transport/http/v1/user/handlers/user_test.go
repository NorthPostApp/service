package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserRouter(handler *UserHandler, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/user/signin", mockAuthMiddleware(uid), handler.AuthenticateAppUser)
	return r
}

func TestAuthenticateAppUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		uid            string
		mockOutput     *models.AppUser
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name: "success",
			uid:  "user-123",
			mockOutput: &models.AppUser{
				Email:       "test@example.com",
				DisplayName: "Test User",
				LastLogin:   123,
				ImageUrl:    "https://example.com/avatar.png",
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
			mockRepo := new(mockUserRepo)
			handler := NewUserHandler(mockRepo, slog.Default())
			router := setupUserRouter(handler, tt.uid)
			if tt.expectCall {
				opts := repository.GetUserByIdOptions{Uid: tt.uid}
				mockRepo.On("AuthenticateAppUserById", mock.Anything, opts).
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
			mockRepo.AssertExpectations(t)
		})
	}
}
