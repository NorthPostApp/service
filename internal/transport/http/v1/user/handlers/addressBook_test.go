package handlers

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAddressBookRouter(handler *AddressBookHandler, uid string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PATCH("/user/addressBook", mockAuthMiddleware(uid), handler.UpdateSavedAddresses)
	return r
}

func TestUpdateSavedAddresses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		uid            string
		body           string
		mockOutput     string
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name:           "success",
			uid:            "mock_user",
			body:           `{"addressId":"test_id","action":"add"}`,
			mockOutput:     "test_id",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectCall:     true,
		},
		{
			name:           "missing uid",
			uid:            "",
			body:           "",
			mockOutput:     "",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectCall:     false,
		},
		{
			name:           "invalid body",
			uid:            "mock_user",
			body:           `{a}`,
			mockOutput:     "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
		{
			name:           "failed service",
			uid:            "mock_user",
			body:           `{"addressId":"test_id","action":"a"}`,
			mockOutput:     "test_id",
			mockError:      errors.New("invalid method"),
			expectedStatus: http.StatusInternalServerError,
			expectCall:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mockUserRepo)
			mockAddressRepo := new(mockAddressRepo)
			handler := NewAddressBookHandler(mockUserRepo,
				mockAddressRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)))
			router := setupAddressBookRouter(handler, tt.uid)
			if tt.expectCall {
				mockUserRepo.On("UpdateUserSavedAddresses", mock.Anything, mock.Anything).
					Return(tt.mockOutput, tt.mockError).Once()
			}
			req, _ := http.NewRequest("PATCH", "/user/addressBook", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectCall {
				mockUserRepo.AssertExpectations(t)
			}
			resp := w.Body.String()
			if tt.expectCall && tt.mockError == nil {
				assert.Contains(t, resp, "data")
				assert.Contains(t, resp, "test_id")
			} else {
				assert.Contains(t, resp, "error")
			}
		})
	}
}
