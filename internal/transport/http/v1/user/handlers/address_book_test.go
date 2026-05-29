package handlers

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"north-post/service/internal/domain/v1/models"
	"north-post/service/internal/repository"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAddressBookRouter(handler *AddressBookHandler, uid string, language string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	group := r.Group("/user/address-book",
		mockAuthMiddleware(uid),
		mockLanguageMiddleware(language),
	)
	group.PATCH("", handler.UpdateSavedAddresses)
	group.GET("", handler.GetSavedAddresses)
	group.POST("/request", handler.CreateNewRequest)
	group.GET("/request", handler.GetRequestsByIDs)
	return r
}

func TestUpdateSavedAddresses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		uid            string
		body           string
		language       string
		mockOutput     string
		mockError      error
		expectedStatus int
		expectCall     bool
	}{
		{
			name:           "success",
			uid:            "mock_user",
			body:           `{"addressIDs":["test_id"],"action":"add"}`,
			language:       "zh",
			mockOutput:     "timestamp",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectCall:     true,
		},
		{
			name:           "missing uid",
			uid:            "",
			body:           "",
			language:       "zh",
			mockOutput:     "",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectCall:     false,
		},
		{
			name:           "invalid body",
			uid:            "mock_user",
			body:           `{a}`,
			language:       "zh",
			mockOutput:     "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectCall:     false,
		},
		{
			name:           "failed service",
			uid:            "mock_user",
			body:           `{"addressIDs":["test_id"],"action":"a"}`,
			language:       "zh",
			mockOutput:     "timestamp",
			mockError:      errors.New("invalid method"),
			expectedStatus: http.StatusInternalServerError,
			expectCall:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mockUserRepo)
			mockAddressRepo := new(mockAddressRepo)
			mockAddressRequestRepo := new(mockAddressRequestRepo)
			handler := NewAddressBookHandler(
				mockUserRepo,
				mockAddressRepo,
				mockAddressRequestRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)))
			router := setupAddressBookRouter(handler, tt.uid, tt.language)
			if tt.expectCall {
				mockUserRepo.On("UpdateUserSavedAddresses", mock.Anything, mock.Anything).
					Return(tt.mockOutput, tt.mockError).Once()
			}
			req, _ := http.NewRequest("PATCH", "/user/address-book", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectCall {
				mockUserRepo.AssertExpectations(t)
			}
			resp := w.Body.String()
			if tt.expectCall && tt.mockError == nil {
				assert.Contains(t, resp, "data")
				assert.Contains(t, resp, "timestamp")
			} else {
				assert.Contains(t, resp, "error")
			}
		})
	}
}

func TestCreateNewRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		uid                         string
		language                    string
		body                        string
		mockCreateRequestOutput     string
		mockCreateRequestError      error
		expectCreateRequestCall     bool
		mockUpdateUserRequestOutput string
		mockUpdateUserRequestError  error
		expectUpdateRequestCall     bool
		expectedStatus              int
	}{
		{
			name:                        "success",
			uid:                         "user_id",
			language:                    "zh",
			body:                        `{"language":"zh","content":"test"}`,
			mockCreateRequestOutput:     "id-123",
			mockCreateRequestError:      nil,
			expectCreateRequestCall:     true,
			mockUpdateUserRequestOutput: "123456",
			mockUpdateUserRequestError:  nil,
			expectUpdateRequestCall:     true,
			expectedStatus:              http.StatusOK,
		},
		{
			name:                        "missing id",
			uid:                         "",
			language:                    "zh",
			body:                        `{"language":"zh","content":"test"}`,
			mockCreateRequestOutput:     "",
			mockCreateRequestError:      nil,
			expectCreateRequestCall:     false,
			mockUpdateUserRequestOutput: "",
			mockUpdateUserRequestError:  nil,
			expectUpdateRequestCall:     false,
			expectedStatus:              http.StatusUnauthorized,
		},
		{
			name:                        "invalid body",
			uid:                         "user_id",
			language:                    "zh",
			body:                        `a`,
			mockCreateRequestOutput:     "",
			mockCreateRequestError:      nil,
			expectCreateRequestCall:     false,
			mockUpdateUserRequestOutput: "",
			mockUpdateUserRequestError:  nil,
			expectUpdateRequestCall:     false,
			expectedStatus:              http.StatusBadRequest,
		},
		{
			name:                        "create call failed",
			uid:                         "user_id",
			language:                    "zh",
			body:                        `{"language":"zh","content":"test"}`,
			mockCreateRequestOutput:     "",
			mockCreateRequestError:      errors.New("error"),
			expectCreateRequestCall:     true,
			mockUpdateUserRequestOutput: "",
			mockUpdateUserRequestError:  nil,
			expectUpdateRequestCall:     false,
			expectedStatus:              http.StatusInternalServerError,
		},
		{
			name:                        "second call failed",
			uid:                         "user_id",
			language:                    "zh",
			body:                        `{"language":"zh","content":"test"}`,
			mockCreateRequestOutput:     "id-123",
			mockCreateRequestError:      nil,
			expectCreateRequestCall:     true,
			mockUpdateUserRequestOutput: "",
			mockUpdateUserRequestError:  errors.New("error"),
			expectUpdateRequestCall:     true,
			expectedStatus:              http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mockUserRepo)
			mockAddressRepo := new(mockAddressRepo)
			mockAddressRequestRepo := new(mockAddressRequestRepo)
			handler := NewAddressBookHandler(
				mockUserRepo,
				mockAddressRepo,
				mockAddressRequestRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)),
			)
			router := setupAddressBookRouter(handler, tt.uid, tt.language)
			if tt.expectCreateRequestCall {
				mockAddressRequestRepo.On("CreateNewRequestWithLimit", mock.Anything, mock.Anything).
					Return(tt.mockCreateRequestOutput, tt.mockCreateRequestError).Once()
			}
			if tt.expectUpdateRequestCall {
				mockUserRepo.On("UpdateUserAddressRequests",
					mock.Anything,
					mock.MatchedBy(func(opts *repository.UpdateUserAddressRequestsOptions) bool {
						return opts.RequestIDs[0] == tt.mockCreateRequestOutput
					}), mock.Anything).
					Return(tt.mockUpdateUserRequestOutput, tt.mockUpdateUserRequestError).Once()
			}
			req, _ := http.NewRequest("POST", "/user/address-book/request", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			resp := w.Body.String()
			if tt.expectUpdateRequestCall && tt.mockUpdateUserRequestError == nil {
				assert.Contains(t, resp, tt.mockCreateRequestOutput)
			}
		})
	}
}

func TestGetSavedAddresses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		uid                     string
		language                string
		getSavedAddressesOutput []string
		getSavedAddressesError  error
		expectGetSavedCall      bool
		getAddressesByIDsOutput *repository.GetAddressesByIDsResponse
		getAddressesByIDsError  error
		expectGetAddressesCall  bool
		expectCleanupCall       bool
		expectedStatus          int
	}{
		{
			name:                    "success",
			uid:                     "mock_user",
			language:                "zh",
			getSavedAddressesOutput: []string{"address-1", "invalid-1"},
			expectGetSavedCall:      true,
			getAddressesByIDsOutput: &repository.GetAddressesByIDsResponse{
				Addresses: []models.AddressItem{
					{
						ID:   "address-1",
						Name: "test address",
					},
				},
				InvalidIDs: []string{"invalid-1"},
			},
			expectGetAddressesCall: true,
			expectCleanupCall:      true,
			expectedStatus:         http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mockUserRepo)
			mockAddressRepo := new(mockAddressRepo)
			mockAddressRequestRepo := new(mockAddressRequestRepo)

			handler := NewAddressBookHandler(
				mockUserRepo,
				mockAddressRepo,
				mockAddressRequestRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)),
			)

			router := setupAddressBookRouter(handler, tt.uid, tt.language)
			if tt.expectGetSavedCall {
				mockUserRepo.On(
					"GetUserSavedAddresses",
					mock.Anything,
					mock.Anything,
				).Return(tt.getSavedAddressesOutput, tt.getSavedAddressesError).Once()
			}
			if tt.expectGetAddressesCall {
				mockAddressRepo.On(
					"GetAddressesByIDs",
					mock.Anything,
					mock.Anything,
				).Return(tt.getAddressesByIDsOutput, tt.getAddressesByIDsError).Once()
			}
			cleanupDone := make(chan *repository.UpdateUserSavedAddressesOptions, 1)
			if tt.expectCleanupCall {
				mockUserRepo.On(
					"UpdateUserSavedAddresses",
					mock.Anything,
					mock.MatchedBy(func(opts *repository.UpdateUserSavedAddressesOptions) bool {
						cleanupDone <- opts
						return opts.UserID == tt.uid &&
							opts.Action == repository.Delete &&
							assert.ElementsMatch(t, []string{"invalid-1"}, opts.AddressIDs)
					}),
				).Return("", nil).Once()
			}

			req, _ := http.NewRequest("GET", "/user/address-book", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectCleanupCall {
				select {
				case opts := <-cleanupDone:
					assert.Equal(t, tt.uid, opts.UserID)
					assert.Equal(t, models.Language(tt.language), opts.Language)
				case <-time.After(time.Second):
					t.Fatal("time out waiting for background cleanup call")
				}
			}
			mockAddressRepo.AssertExpectations(t)
		})
	}
}

func TestGetRequestsByIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		uid                         string
		language                    string
		expectGetUserRequestsCall   bool
		getUserRequestsOutput       *models.AddressRequests
		getUserRequestsError        error
		expectedGetRequestByIDsCall bool
		getRequestsByIDsOutput      *repository.GetRequestsByIDsResponse
		getRequestsByIDsError       error
		expectAsyncRemoveIDsCall    bool
		expectAsyncUpdateCountCall  bool
		expectedStatus              int
	}{
		{
			name:                      "success",
			uid:                       "user_id",
			language:                  "en",
			expectGetUserRequestsCall: true,
			getUserRequestsOutput: &models.AddressRequests{
				IDs:                []string{"id_1", "id_2"},
				ActiveRequestCount: 2},
			getUserRequestsError:        nil,
			expectedGetRequestByIDsCall: true,
			getRequestsByIDsOutput: &repository.GetRequestsByIDsResponse{
				Requests:            []models.AddressRequest{{ID: "123"}},
				ActiveRequestsCount: 2,
			},
			expectAsyncRemoveIDsCall:   false,
			expectAsyncUpdateCountCall: false,
			expectedStatus:             http.StatusOK,
		},
		{
			name:                      "success with remove invalid ids",
			uid:                       "user_id",
			language:                  "en",
			expectGetUserRequestsCall: true,
			getUserRequestsOutput: &models.AddressRequests{
				IDs:                []string{"id_1", "id_2"},
				ActiveRequestCount: 2},
			getUserRequestsError:        nil,
			expectedGetRequestByIDsCall: true,
			getRequestsByIDsOutput: &repository.GetRequestsByIDsResponse{
				Requests:            []models.AddressRequest{{ID: "123"}},
				InvalidIDs:          []string{"invalid-1"},
				ActiveRequestsCount: 2,
			},
			expectAsyncRemoveIDsCall:   true,
			expectAsyncUpdateCountCall: false,
			expectedStatus:             http.StatusOK,
		},
		{
			name:                      "success with new counts",
			uid:                       "user_id",
			language:                  "en",
			expectGetUserRequestsCall: true,
			getUserRequestsOutput: &models.AddressRequests{
				IDs:                []string{"id_1", "id_2"},
				ActiveRequestCount: 2},
			getUserRequestsError:        nil,
			expectedGetRequestByIDsCall: true,
			getRequestsByIDsOutput: &repository.GetRequestsByIDsResponse{
				Requests:            []models.AddressRequest{{ID: "123"}},
				InvalidIDs:          []string{"invalid-1"},
				ActiveRequestsCount: 4,
			},
			expectAsyncRemoveIDsCall:   true,
			expectAsyncUpdateCountCall: true,
			expectedStatus:             http.StatusOK,
		},
		{
			name:                        "invalid id",
			uid:                         "",
			language:                    "en",
			expectGetUserRequestsCall:   false,
			getUserRequestsOutput:       nil,
			getUserRequestsError:        nil,
			expectedGetRequestByIDsCall: false,
			getRequestsByIDsOutput:      nil,
			expectAsyncRemoveIDsCall:    false,
			expectAsyncUpdateCountCall:  false,
			expectedStatus:              http.StatusUnauthorized,
		},
		{
			name:                        "failed GetUserRequests",
			uid:                         "user_id",
			language:                    "en",
			expectGetUserRequestsCall:   true,
			getUserRequestsOutput:       nil,
			getUserRequestsError:        errors.New("failed call"),
			expectedGetRequestByIDsCall: false,
			getRequestsByIDsOutput:      nil,
			expectAsyncRemoveIDsCall:    false,
			expectAsyncUpdateCountCall:  false,
			expectedStatus:              http.StatusInternalServerError,
		},
		{
			name:                      "failed GetRequestsByIDs",
			uid:                       "user_id",
			language:                  "en",
			expectGetUserRequestsCall: true,
			getUserRequestsOutput: &models.AddressRequests{
				IDs:                []string{"id_1", "id_2"},
				ActiveRequestCount: 2},
			getUserRequestsError:        nil,
			expectedGetRequestByIDsCall: true,
			getRequestsByIDsOutput:      nil,
			getRequestsByIDsError:       errors.New("failed"),
			expectAsyncRemoveIDsCall:    false,
			expectAsyncUpdateCountCall:  false,
			expectedStatus:              http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mockUserRepo)
			mockAddressRepo := new(mockAddressRepo)
			mockAddressRequestRepo := new(mockAddressRequestRepo)
			handler := NewAddressBookHandler(
				mockUserRepo,
				mockAddressRepo,
				mockAddressRequestRepo,
				slog.New(slog.NewTextHandler(io.Discard, nil)))
			router := setupAddressBookRouter(handler, tt.uid, tt.language)
			if tt.expectGetUserRequestsCall {
				mockUserRepo.On(
					"GetUserRequests",
					mock.Anything,
					mock.Anything,
				).Return(tt.getUserRequestsOutput, tt.getUserRequestsError).Once()
			}
			if tt.expectedGetRequestByIDsCall {
				mockAddressRequestRepo.On(
					"GetRequestsByIDs",
					mock.Anything,
					mock.Anything,
				).Return(tt.getRequestsByIDsOutput, tt.getRequestsByIDsError).Once()
			}

			invalidIDsCleanup := make(chan *repository.UpdateUserAddressRequestsOptions, 1)
			activeCountUpdate := make(chan *repository.UpdateUserActiveRequestCountOptions, 1)

			if tt.expectAsyncRemoveIDsCall {
				mockUserRepo.On(
					"UpdateUserAddressRequests",
					mock.Anything,
					mock.MatchedBy(func(opts *repository.UpdateUserAddressRequestsOptions) bool {
						invalidIDsCleanup <- opts
						return opts.UserID == tt.uid &&
							opts.Language == models.Language(tt.language) &&
							opts.Action == repository.Delete &&
							assert.ElementsMatch(t, tt.getRequestsByIDsOutput.InvalidIDs, opts.RequestIDs)
					}),
				).Return("", nil).Once()
			}

			if tt.expectAsyncUpdateCountCall {
				mockUserRepo.On(
					"UpdateUserActiveRequestCount",
					mock.Anything,
					mock.MatchedBy(func(opts *repository.UpdateUserActiveRequestCountOptions) bool {
						activeCountUpdate <- opts
						return opts.Count == tt.getRequestsByIDsOutput.ActiveRequestsCount
					}),
				).Return(nil).Once()
			}

			req, _ := http.NewRequest("GET", "/user/address-book/request", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if tt.expectAsyncRemoveIDsCall {
				select {
				case opts := <-invalidIDsCleanup:
					assert.Equal(t, tt.uid, opts.UserID)
					assert.Equal(t, tt.language, opts.Language.Get())
				case <-time.After(time.Second):
					t.Fatal("time out waiting for background clean up call")
				}
			}
			if tt.expectAsyncUpdateCountCall {
				select {
				case opts := <-activeCountUpdate:
					assert.Equal(t, tt.getRequestsByIDsOutput.ActiveRequestsCount, opts.Count)
				case <-time.After(time.Second):
					t.Fatal("time out waiting for background update call")
				}
			}
			mockUserRepo.AssertExpectations(t)
			mockAddressRepo.AssertExpectations(t)
		})
	}
}
