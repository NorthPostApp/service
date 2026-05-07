package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupLanguageFromQueryTestContext(queryString string) (
	*gin.Context,
	*httptest.ResponseRecorder,
	*bytes.Buffer,
) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", fmt.Sprintf("/test?%s", queryString), nil)
	c.Request = req
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	middleware := LanguageFromQueryMiddleware(logger)
	middleware(c)
	return c, w, &logBuffer
}

func setupLanguageFromBodyTestContext(body string) (
	*gin.Context,
	*httptest.ResponseRecorder,
	*bytes.Buffer,
) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	middleware := LanguageFromBodyMiddleware(logger)
	middleware(c)
	return c, w, &logBuffer
}

func TestLanguageFromQueryMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedValue string
		status        int
		abort         bool
		errorMessage  string
	}{
		{
			name:          "success",
			query:         "language=zh",
			expectedValue: "zh",
			status:        http.StatusOK,
			abort:         false,
			errorMessage:  "",
		},
		{
			name:          "missing language query",
			query:         "",
			expectedValue: "",
			status:        http.StatusBadRequest,
			abort:         true,
			errorMessage:  "language query is required",
		},
		{
			name:          "invalid language",
			query:         "language=abc",
			expectedValue: "",
			status:        http.StatusBadRequest,
			abort:         true,
			errorMessage:  "invalid language",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w, logBuffer := setupLanguageFromQueryTestContext(tt.query)
			defer teardownTest()
			logOutput := logBuffer.String()
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, tt.abort, c.IsAborted())
			assert.Contains(t, logOutput, tt.errorMessage)
			if !tt.abort {
				assert.Equal(t, c.MustGet(LanguageKey), tt.expectedValue)
			}
		})
	}
}

func TestLanguageFromBodyMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedValue string
		status        int
		abort         bool
		errorMessage  string
	}{
		{
			name:          "success",
			body:          `{"language":"en"}`,
			expectedValue: "en",
			status:        http.StatusOK,
			abort:         false,
			errorMessage:  "",
		},
		{
			name:          "missing language body",
			body:          "",
			expectedValue: "",
			status:        http.StatusBadRequest,
			abort:         true,
			errorMessage:  "failed to bind",
		},
		{
			name:          "invalid language",
			body:          `{"language":"abc"}`,
			expectedValue: "",
			status:        http.StatusBadRequest,
			abort:         true,
			errorMessage:  "invalid language",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w, logBuffer := setupLanguageFromBodyTestContext(tt.body)
			defer teardownTest()
			logOutput := logBuffer.String()
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, tt.abort, c.IsAborted())
			assert.Contains(t, logOutput, tt.errorMessage)
			if !tt.abort {
				assert.Equal(t, c.MustGet(LanguageKey), tt.expectedValue)
			}
		})
	}
}
