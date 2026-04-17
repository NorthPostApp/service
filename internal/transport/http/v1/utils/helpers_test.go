package utils

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type TestRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func TestBindJSON(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		expectedResult bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "valid JSON request",
			body: TestRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedResult: true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid JSON - missing required field",
			body:           map[string]string{"name": "John"},
			expectedResult: false,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "invalid email format",
			body: map[string]string{
				"name":  "John Doe",
				"email": "not-an-email",
			},
			expectedResult: false,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			bodyBytes, _ := json.Marshal(tt.body)
			c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			var req TestRequest
			result := BindJSON(c, &req, logger)
			assert.Equal(t, result, tt.expectedResult)
			if tt.expectedError {
				assert.Equal(t, tt.expectedStatus, w.Code)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}
			if tt.expectedResult {
				assert.Equal(t, req.Email, "john@example.com")
				assert.Equal(t, req.Name, "John Doe")
			}
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		name           string
		language       models.Language
		expectedResult bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "valid language",
			language:       models.LanguageZH,
			expectedResult: true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid language",
			language:       models.Language("abc"),
			expectedResult: false,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			result := ValidateLanguage(c, tt.language, logger)
			assert.Equal(t, result, tt.expectedResult)
			if tt.expectedError {
				assert.Equal(t, tt.expectedStatus, w.Code)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.Nil(t, err)
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestValidateMusicFilename(t *testing.T) {
	tests := []struct {
		name           string
		genre          string
		track          string
		expectedResult bool
	}{
		{"valid inputs", "pop", "song.mp3", true},
		{"empty genre", "", "song.mp3", false},
		{"empty track", "pop", "", false},
		{"path traversal in genre", "../../etc", "passwd", false},
		{"path traversal in track", "pop", "../secret.mp3", false},
		{"slash in track", "pop", "sub/song.mp3", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
			result := ValidateMusicFilename(c, tt.genre, tt.track, logger)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestStringToFloat32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float32
	}{
		{
			name:     "success",
			input:    "32.0",
			expected: float32(32.0),
		},
		{
			name:     "empty string",
			input:    "",
			expected: float32(0),
		},
		{
			name:     "invalid value",
			input:    "invalid",
			expected: float32(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := StringToFloat32(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestStringToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "success",
			input:    "32.0",
			expected: int64(32),
		},
		{
			name:     "empty string",
			input:    "",
			expected: int64(0),
		},
		{
			name:     "invalid value",
			input:    "invalid",
			expected: int64(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := StringToInt64(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}
