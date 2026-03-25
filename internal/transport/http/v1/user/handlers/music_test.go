package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMusicService struct {
	mock.Mock
}

func (m *mockMusicService) GetAllMusicList(ctx context.Context) (
	*services.GetAllMusicListOutput,
	error,
) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.GetAllMusicListOutput), args.Error(1)
}

func (m *mockMusicService) GetPresignedMusicURL(
	ctx context.Context,
	input services.GetPresignedMusicURLInput,
) (*services.GetPresignedMusicURLOutput, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.GetPresignedMusicURLOutput), args.Error(1)
}

func setupMusicRouter() (*mockMusicService, *gin.Engine) {
	mockSrv := new(mockMusicService)
	handler := NewMusicHandler(mockSrv, slog.Default())
	router := gin.Default()
	router.GET("/user/music/list", handler.GetMusicList)
	router.GET("/user/music/:genre/:track", handler.GetPresignedMusicURL)
	return mockSrv, router
}

func TestMusicHandler_GetMusicList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.GetAllMusicListOutput
		mockError      error
		expectedStatus int
	}{
		{

			name: "success request",
			url:  "/user/music/list",
			mockOutput: &services.GetAllMusicListOutput{
				Data: []models.Music{
					{Filename: "test"},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			url:            "/user/music/list",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv, router := setupMusicRouter()
			mockSrv.On("GetAllMusicList", mock.Anything).Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockOutput == nil && tt.mockError == nil {
				mockSrv.AssertNotCalled(t, "GetAllMusicList")
			} else if tt.mockError != nil {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}

func TestMusicHandler_GetPresignedMusicURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.GetPresignedMusicURLOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success request",
			url:            "/user/music/foo/bar.mp3",
			mockOutput:     &services.GetPresignedMusicURLOutput{URL: "foo/bar.mp3"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request with empty genre",
			url:            "/user/music//bar.mp3",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request with empty track name",
			url:            "/user/music/a/ ",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request with possible sourcing name",
			url:            "/user/music/../a",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "failed request",
			url:            "/user/music/a/b.mp3",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrc, router := setupMusicRouter()
			mockSrc.On("GetPresignedMusicURL",
				mock.Anything,
				mock.Anything).Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockError != nil && tt.expectedStatus != http.StatusBadRequest {
				mockSrc.AssertExpectations(t)
			}
		})
	}
}
