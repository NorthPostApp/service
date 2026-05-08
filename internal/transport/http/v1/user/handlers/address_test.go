package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"north-post/service/internal/transport/http/v1/dto"
	"north-post/service/internal/transport/http/v1/middleware"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAddressService struct {
	mock.Mock
}

func (m *MockAddressService) GetAllTags(
	ctx context.Context,
	input services.GetAllTagsInput) (*services.GetAllTagsOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.GetAllTagsOutput), args.Error(1)
}

func setupRouter(handler *AddressHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r.GET("/user/address/tags", middleware.LanguageFromQueryMiddleware(logger), handler.GetAllTags)
	return r
}

func TestGetAllTags(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		language       models.Language
		url            string
		mockOutput     *services.GetAllTagsOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			language: "EN",
			url:      "/user/address/tags?language=en",
			mockOutput: &services.GetAllTagsOutput{
				TagsRecord: models.TagsRecord{
					Tags:        map[string][]string{"test": {"test1", "test2"}},
					RefreshedAt: 123,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service returns error",
			language:       "ZH",
			url:            "/user/address/tags?language=zH",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid language",
			language:       "EN",
			url:            "/user/address/tags?language=abs",
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv.On("GetAllTags", mock.Anything, mock.Anything).
				Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockError == nil && tt.mockOutput != nil {
				var response dto.GetAllTagsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.language.Get(), response.Data.Language.Get())
				assert.Equal(t, tt.mockOutput.TagsRecord.Tags, response.Data.Tags)
				mockSrv.AssertExpectations(t)
			}
		})
	}
}
