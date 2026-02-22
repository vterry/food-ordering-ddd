package rest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ===== Mock =====

type mockRestaurantService struct {
	mock.Mock
}

func (m *mockRestaurantService) CreateRestaurant(ctx context.Context, req input.CreateRestaurantRequest) (*input.RestaurantResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.RestaurantResponse), args.Error(1)
}

func (m *mockRestaurantService) GetRestaurant(ctx context.Context, restId valueobjects.RestaurantID) (*input.RestaurantResponse, error) {
	args := m.Called(ctx, restId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.RestaurantResponse), args.Error(1)
}

func (m *mockRestaurantService) OpenRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error {
	args := m.Called(ctx, restId)
	return args.Error(0)
}

func (m *mockRestaurantService) CloseRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error {
	args := m.Called(ctx, restId)
	return args.Error(0)
}

// ===== Helpers =====

func setupRestaurantHandler(svc *mockRestaurantService) *http.ServeMux {
	handler := NewRestaurantHandler(svc, slog.Default())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}

// ===== Tests =====

func TestRestaurantHandler_CreateRestaurant(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *mockRestaurantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "should create restaurant and return 201",
			body: `{"name":"Pizza Place","address":{"street":"Rua A","number":"123","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000000"}}`,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("CreateRestaurant", mock.Anything, mock.AnythingOfType("CreateRestaurantRequest")).
					Return(&input.RestaurantResponse{
						ID: "uuid-123", Name: "Pizza Place", Status: "CLOSED",
						Address: input.AddressResponse{Street: "Rua A", Number: "123", Neighborhood: "Centro", City: "SP", State: "SP", ZipCode: "00000000"},
					}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"name":"Pizza Place"`,
		},
		{
			name:           "should return 400 when body is malformed JSON",
			body:           `{invalid json`,
			mockSetup:      func(svc *mockRestaurantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when required fields are missing",
			body:           `{"name":""}`,
			mockSetup:      func(svc *mockRestaurantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "should return 500 when service fails with unexpected error",
			body: `{"name":"Pizza Place","address":{"street":"Rua A","number":"123","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000000"}}`,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("CreateRestaurant", mock.Anything, mock.Anything).
					Return(nil, errors.New("db connection failed")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockRestaurantService)
			tt.mockSetup(svc)
			mux := setupRestaurantHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/restaurant", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestRestaurantHandler_GetRestaurant(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockRestaurantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should return restaurant and 200",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("GetRestaurant", mock.Anything, mock.Anything).
					Return(&input.RestaurantResponse{
						ID: validUUID, Name: "Pizza Place", Status: "CLOSED",
					}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"name":"Pizza Place"`,
		},
		{
			name:           "should return error when ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockRestaurantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 404 when restaurant not found",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("GetRestaurant", mock.Anything, mock.Anything).
					Return(nil, common.NewNotFoundErr(errors.New("not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "should return 500 when service fails with unexpected error",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("GetRestaurant", mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockRestaurantService)
			tt.mockSetup(svc)
			mux := setupRestaurantHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/restaurant/"+tt.pathID, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestRestaurantHandler_OpenRestaurant(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockRestaurantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should open restaurant and return 200",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("OpenRestaurant", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"restaurant opened"`,
		},
		{
			name:           "should return error when ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockRestaurantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 422 when restaurant is already opened",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("OpenRestaurant", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(restaurant.ErrAlreadyOpened)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"restaurant is already opened"`,
		},
		{
			name:   "should return 422 when restaurant has no active menu",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("OpenRestaurant", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(restaurant.ErrNoActiveMenu)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"missing an active menu"`,
		},
		{
			name:   "should return 500 when service fails with unexpected error",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("OpenRestaurant", mock.Anything, mock.Anything).
					Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockRestaurantService)
			tt.mockSetup(svc)
			mux := setupRestaurantHandler(svc)

			req := httptest.NewRequest(http.MethodPatch, "/restaurant/"+tt.pathID+"/open", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestRestaurantHandler_CloseRestaurant(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockRestaurantService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should close restaurant and return 200",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("CloseRestaurant", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"restaurant closed"`,
		},
		{
			name:           "should return error when ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockRestaurantService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 422 when restaurant is already closed",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("CloseRestaurant", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(restaurant.ErrAlreadyClosed)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"restaurant is already closed"`,
		},
		{
			name:   "should return 500 when service fails with unexpected error",
			pathID: validUUID,
			mockSetup: func(svc *mockRestaurantService) {
				svc.On("CloseRestaurant", mock.Anything, mock.Anything).
					Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockRestaurantService)
			tt.mockSetup(svc)
			mux := setupRestaurantHandler(svc)

			req := httptest.NewRequest(http.MethodPatch, "/restaurant/"+tt.pathID+"/close", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
		})
	}
}
