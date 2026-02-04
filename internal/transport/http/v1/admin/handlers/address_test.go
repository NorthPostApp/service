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

// MockAddressService implements the methods used by AddressHandler for testing.
type MockAddressService struct {
	mock.Mock
}

func (m *MockAddressService) GetAllAddresses(
	ctx context.Context,
	input services.GetAllAddressesInput,
) (*services.GetAllAddressesOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.GetAllAddressesOutput), args.Error(1)
}
func (m *MockAddressService) GetAddressById(
	ctx context.Context,
	input services.GetAddressByIdInput,
) (*services.GetAddressByIdOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.GetAddressByIdOutput), args.Error(1)
}
func (m *MockAddressService) CreateNewAddress(
	ctx context.Context,
	input services.CreateNewAddressInput,
) (*services.CreateNewAddressOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.CreateNewAddressOutput), args.Error(1)
}
func (m *MockAddressService) UpdateAddress(
	ctx context.Context,
	input services.UpdateAddressInput,
) (*services.UpdateAddressOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.UpdateAddressOutput), args.Error(1)
}
func (m *MockAddressService) GenerateNewAddress(
	ctx context.Context,
	input services.GenerateAddressInput,
) (*services.GenerateAddressOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.GenerateAddressOutput), args.Error(1)
}
func (m *MockAddressService) DeleteAddress(
	ctx context.Context,
	input services.DeleteAddressInput,
) (*services.DeleteAddressOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.DeleteAddressOutput), args.Error(1)
}
func (m *MockAddressService) RefreshTags(
	ctx context.Context,
	input services.RefreshTagsInput,
) (*services.RefreshTagsOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.RefreshTagsOutput), args.Error(1)
}
func (m *MockAddressService) GetAllTags(
	ctx context.Context,
	input services.GetAllTagsInput,
) (*services.GetAllTagsOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*services.GetAllTagsOutput), args.Error(1)
}

func setupRouter(handler *AddressHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/admin/address", handler.GetAllAddresses)
	r.GET("/admin/address/:id", handler.GetAddressById)
	r.GET("/admin/address/tags", handler.GetAllTags)
	r.DELETE("/admin/address/:id", handler.DeleteAddress)
	r.PUT("/admin/address", handler.CreateNewAddress)
	r.POST("/admin/address/generate", handler.GenerateNewAddress)
	r.POST("/admin/address/update", handler.UpdateAddress)
	return r
}

func TestGetAllAddresses(t *testing.T) {
	t.Parallel()
	mockSvc := new(MockAddressService)
	handler := NewAddressHandler(mockSvc, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		language       string
		mockOutput     *services.GetAllAddressesOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			language: "en",
			mockOutput: &services.GetAllAddressesOutput{
				Addresses:  []models.AddressItem{{ID: "1"}},
				TotalCount: 1,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			language:       "en",
			mockOutput:     &services.GetAllAddressesOutput{},
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing language",
			language:       "",
			mockOutput:     nil,
			mockError:      errors.New("missing language"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid language",
			language:       "abc",
			mockOutput:     nil,
			mockError:      errors.New("missing language"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := dto.GetAllAddressesRequest{Language: models.Language(tt.language)}
			input := services.GetAllAddressesInput{Language: models.Language(tt.language)}
			mockSvc.On("GetAllAddresses", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/admin/address", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code == http.StatusOK {
				var response dto.GetAllAddressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockOutput.Addresses[0].ID, response.Data.Addresses[0].ID)
				assert.Equal(t, tt.mockOutput.TotalCount, response.Data.TotalCount)
			}
		})
	}
	// one more test for missing request body
	t.Run("bad request body", func(t *testing.T) {
		body, _ := json.Marshal("")
		req, _ := http.NewRequest("POST", "/admin/address", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetAddressesById(t *testing.T) {
	t.Parallel()
	mockSvc := new(MockAddressService)
	handler := NewAddressHandler(mockSvc, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.GetAddressByIdOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			url:            "/admin/address/1?language=en",
			mockOutput:     &services.GetAddressByIdOutput{Address: models.AddressItem{}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			url:            "/admin/address/1?language=en",
			mockOutput:     nil,
			mockError:      errors.New("fail"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing id",
			url:            "/admin/address/?language=en",
			mockOutput:     nil,
			mockError:      errors.New("id required"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty id",
			url:            "/admin/address/ ?language=en",
			mockOutput:     nil,
			mockError:      errors.New("id required"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing language",
			url:            "/admin/address/1",
			mockOutput:     nil,
			mockError:      errors.New("language is requires"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid language",
			url:            "/admin/address/1?language=k",
			mockOutput:     nil,
			mockError:      errors.New("error"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := services.GetAddressByIdInput{Language: "en", ID: "1"}
			mockSvc.On("GetAddressById", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockOutput != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestCreateNewAddress(t *testing.T) {
	t.Parallel()
	mockSvc := new(MockAddressService)
	handler := NewAddressHandler(mockSvc, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		language       string
		mockOutput     *services.CreateNewAddressOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			language:       "en",
			mockOutput:     &services.CreateNewAddressOutput{ID: "1"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			language:       "en",
			mockOutput:     &services.CreateNewAddressOutput{ID: "1"},
			mockError:      errors.New("error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid language",
			language:       "abc",
			mockOutput:     nil,
			mockError:      errors.New("error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing body",
			language:       "",
			mockOutput:     nil,
			mockError:      errors.New("error"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			language := models.Language(tt.language)
			reqBody := dto.CreateAddressRequest{
				Language:   language,
				Name:       "Test Address",
				BriefIntro: "Test intro",
				Tags:       []string{"a", "b"},
				Address: dto.AddressDTO{
					City:    "test",
					Country: "test",
					Line1:   "test",
					Region:  "test",
				},
			}
			input := services.CreateNewAddressInput{Language: language, Address: dto.FromCreateAddressDTO(reqBody)}
			mockSvc.On("CreateNewAddress", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("PUT", "/admin/address", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockOutput != nil {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestGenerateNewAddress(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		body           dto.GenerateNewAddressRequest
		mockOutput     *services.GenerateAddressOutput
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name: "success",
			body: dto.GenerateNewAddressRequest{
				Language: "en",
				Prompt:   "sys",
			},
			mockOutput: &services.GenerateAddressOutput{
				Addresses: []models.AddressItem{
					{ID: "1", Name: "Test Address"},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectCall:     true,
		},
		{
			name: "service error",
			body: dto.GenerateNewAddressRequest{
				Language: "en",
				Prompt:   "sys",
			},
			mockOutput:     nil,
			mockError:      errors.New("service failed"),
			expectedStatus: http.StatusInternalServerError,
			expectCall:     true,
		},
		{
			name:           "invalid json",
			body:           dto.GenerateNewAddressRequest{},
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
		{
			name: "invalid language",
			body: dto.GenerateNewAddressRequest{
				Language: "k",
				Prompt:   "sys",
			},
			mockOutput:     nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := services.GenerateAddressInput{
				Language: tt.body.Language,
				Prompt:   tt.body.Prompt,
			}
			mockSrv.On("GenerateNewAddress", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/admin/address/generate", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectCall {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}

func TestUpdateAddress(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	mockAddressItem := dto.AddressItemDTO{
		ID:         "1",
		Name:       "Test Address",
		BriefIntro: "Test intro",
		Tags:       []string{"a", "b"},
		Address: dto.AddressDTO{
			City:    "test",
			Country: "test",
			Line1:   "test",
			Region:  "test",
		},
	}
	tests := []struct {
		name           string
		body           dto.UpdateAddressRequest
		mockOutput     *services.UpdateAddressOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			body:           dto.UpdateAddressRequest{Language: "EN", ID: "1", Address: mockAddressItem},
			mockOutput:     &services.UpdateAddressOutput{Address: models.AddressItem{ID: "1"}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			body:           dto.UpdateAddressRequest{Language: "EN", ID: "1", Address: mockAddressItem},
			mockOutput:     nil,
			mockError:      errors.New("update failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid language",
			body:           dto.UpdateAddressRequest{Language: "abc", ID: "1", Address: mockAddressItem},
			mockOutput:     nil,
			mockError:      errors.New("update failed"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing items in request body",
			body:           dto.UpdateAddressRequest{Address: mockAddressItem},
			mockOutput:     nil,
			mockError:      errors.New("update failed"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := services.UpdateAddressInput{
				Language: tt.body.Language,
				ID:       tt.body.ID,
				Address:  dto.FromUpdateAddressDTO(tt.body),
			}
			mockSrv.On("UpdateAddress", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/admin/address/update", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockOutput != nil {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}

func TestDeleteAddress(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.DeleteAddressOutput
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success",
			url:            "/admin/address/1?language=en",
			mockOutput:     &services.DeleteAddressOutput{ID: "1"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			url:            "/admin/address/1?language=en",
			mockOutput:     nil,
			mockError:      errors.New("failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing id",
			url:            "/admin/address/ ?language=en",
			mockOutput:     nil,
			mockError:      errors.New("id required"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing language",
			url:            "/admin/address/1",
			mockOutput:     nil,
			mockError:      errors.New("language required"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid language",
			url:            "/admin/address/1?language=abc",
			mockOutput:     nil,
			mockError:      errors.New("language required"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := services.DeleteAddressInput{Language: "en", ID: "1"}
			mockSrv.On("DeleteAddress", mock.Anything, input).
				Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("DELETE", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockOutput != nil {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}

func TestGetAllTags_RefreshTags(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.RefreshTagsOutput
		mockError      error
		expectedStatus int
	}{
		{
			name: "success",
			url:  "/admin/address/tags?language=zh&refresh=true",
			mockOutput: &services.RefreshTagsOutput{TagsRecord: models.TagsRecord{
				RefreshedAt: 123,
				Tags:        map[string][]string{"test": {"test", "test2"}}},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			url:            "/admin/address/tags?language=zh&refresh=true",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing language",
			url:            "/admin/address/tags",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid language",
			url:            "/admin/address/tags?language=abc",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid refresh query",
			url:            "/admin/address/tags?language=en&refresh=abc",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv.On("RefreshTags", mock.Anything, mock.Anything).
				Return(tt.mockOutput, tt.mockError).Once()
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.mockError == nil {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}

func TestGetAllTags_GetAllTags(t *testing.T) {
	t.Parallel()
	mockSrv := new(MockAddressService)
	handler := NewAddressHandler(mockSrv, slog.Default())
	router := setupRouter(handler)
	tests := []struct {
		name           string
		url            string
		mockOutput     *services.GetAllTagsOutput
		mockError      error
		expectedStatus int
	}{
		{
			name: "success with refresh parameter",
			url:  "/admin/address/tags?language=zh&refresh=false",
			mockOutput: &services.GetAllTagsOutput{TagsRecord: models.TagsRecord{
				RefreshedAt: 123,
				Tags:        map[string][]string{"test": {"test", "test2"}}},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "success without refresh parameter",
			url:  "/admin/address/tags?language=zh",
			mockOutput: &services.GetAllTagsOutput{TagsRecord: models.TagsRecord{
				RefreshedAt: 123,
				Tags:        map[string][]string{"test": {"test", "test2"}}},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			url:            "/admin/address/tags?language=zh",
			mockOutput:     nil,
			mockError:      errors.New("failed request"),
			expectedStatus: http.StatusInternalServerError,
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
			if tt.mockError == nil {
				mockSrv.AssertExpectations(t)
			}
		})
	}
}
