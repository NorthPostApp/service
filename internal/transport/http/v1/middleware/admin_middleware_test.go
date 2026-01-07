package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockAuthClient implements Auth for testing
type MockAuthClient struct {
	VerifyIDTokenFn func(c context.Context, idToken string) (*auth.Token, error)
}

func (m *MockAuthClient) VerifyIDToken(c context.Context, idToken string) (*auth.Token, error) {
	return m.VerifyIDTokenFn(c, idToken)
}

func setupTestContext(authHeader string, mockAuth *MockAuthClient) (
	*gin.Context,
	*httptest.ResponseRecorder,
	*bytes.Buffer,
	gin.HandlerFunc,
) {
	w := httptest.NewRecorder() // w stands for writer/recorder
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/test", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	c.Request = req
	// capture log output
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	// var authClient *auth.Client
	middleware := AdminAuthMiddleware(mockAuth, logger)
	return c, w, &logBuffer, middleware
}

func teardownTest() {
	log.SetOutput(os.Stderr)
}

func TestAdminAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	mockAuth := &MockAuthClient{}
	c, w, logBuffer, middleware := setupTestContext("", mockAuth)
	defer teardownTest()
	middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header required.", response["error"])
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Invalid admin request")
	assert.Contains(t, logOutput, "Authorization header required.")
}

func TestAdminAuthMiddleware_InvalidFormat_NoBearerPrefix(t *testing.T) {
	mockAuth := &MockAuthClient{}
	c, w, logBuffer, middleware := setupTestContext("InvalidToken123", mockAuth)
	defer teardownTest()
	middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Authorization header format must be Bearer")
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Invalid authorization header")
}

func TestAdminAuthMiddleware_InvalidHeader(t *testing.T) {
	mockAuth := &MockAuthClient{}
	c, w, logBuffer, middleware := setupTestContext("Basic token123", mockAuth)
	defer teardownTest()
	middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Authorization header format must be Bearer")
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Invalid authorization header")
}

func TestAdminAuthMiddleware_ValidFormat(t *testing.T) {
	mockAuth := &MockAuthClient{
		VerifyIDTokenFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
			return &auth.Token{
				UID: "test-user-123",
				Claims: map[string]interface{}{
					"email": "test@example.com",
				},
			}, nil
		},
	}
	c, w, logBuffer, middleware := setupTestContext("Bearer valid_token", mockAuth)
	defer teardownTest()
	middleware(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, c.IsAborted())
	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "Invalid admin request")
	assert.NotContains(t, logOutput, "Invalid authorization header")
}

func TestAdminAuthMiddleware_Unauthorized(t *testing.T) {
	mockAuth := &MockAuthClient{
		VerifyIDTokenFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
			return nil, errors.New("Unauthorized")
		},
	}
	c, w, logBuffer, middleware := setupTestContext("Bearer valid_token", mockAuth)
	defer teardownTest()
	middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Unauthorized")
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Unauthorized")
}
