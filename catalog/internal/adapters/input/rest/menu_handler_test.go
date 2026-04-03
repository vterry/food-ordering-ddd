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
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ===== Mock =====

type mockMenuService struct {
	mock.Mock
}

func (m *mockMenuService) CreateMenu(ctx context.Context, restId valueobjects.RestaurantID, req input.CreateMenuRequest) (*input.MenuResponse, error) {
	args := m.Called(ctx, restId, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.MenuResponse), args.Error(1)
}

func (m *mockMenuService) GetMenu(ctx context.Context, menuId valueobjects.MenuID) (*input.MenuResponse, error) {
	args := m.Called(ctx, menuId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.MenuResponse), args.Error(1)
}

type mockCatalogQueryService struct {
	mock.Mock
}

func (m *mockCatalogQueryService) ValidateOrder(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.ValidateOrderResponse), args.Error(1)
}

func (m *mockCatalogQueryService) GetActiveMenu(ctx context.Context, restaurantId string) (*input.MenuResponse, error) {
	args := m.Called(ctx, restaurantId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.MenuResponse), args.Error(1)
}

func (m *mockMenuService) ActiveMenu(ctx context.Context, menuId valueobjects.MenuID) error {
	args := m.Called(ctx, menuId)
	return args.Error(0)
}

func (m *mockMenuService) ArchiveMenu(ctx context.Context, menuId valueobjects.MenuID) error {
	args := m.Called(ctx, menuId)
	return args.Error(0)
}

func (m *mockMenuService) AddCategory(ctx context.Context, menuId valueobjects.MenuID, req input.AddCategoryRequest) error {
	args := m.Called(ctx, menuId, req)
	return args.Error(0)
}

func (m *mockMenuService) AddItemToCategory(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, req input.AddItemRequest) error {
	args := m.Called(ctx, menuId, categoryId, req)
	return args.Error(0)
}

func (m *mockMenuService) UpdateItem(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, req input.UpdateItemRequest) error {
	args := m.Called(ctx, menuId, categoryId, itemId, req)
	return args.Error(0)
}

// ===== Helpers =====

func setupMenuHandler(svc *mockMenuService, qSvc *mockCatalogQueryService) *http.ServeMux {
	handler := NewMenuHandler(svc, qSvc, slog.Default())
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return mux
}

// ===== Tests =====

func TestMenuHandler_CreateMenu(t *testing.T) {
	validRestUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		body           string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should create menu and return 201",
			pathID: validRestUUID,
			body:   `{"name":"Menu Principal"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("CreateMenu", mock.Anything, mock.Anything, mock.AnythingOfType("CreateMenuRequest")).
					Return(&input.MenuResponse{
						ID: "menu-uuid", Name: "Menu Principal", RestaurantID: validRestUUID, Status: "DRAFT",
						Categories: []*input.CategoryResponse{},
					}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"name":"Menu Principal"`,
		},
		{
			name:           "should return error when restaurant ID is not a valid UUID",
			pathID:         "not-a-uuid",
			body:           `{"name":"Menu Principal"}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when body is malformed JSON",
			pathID:         validRestUUID,
			body:           `{invalid`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when required fields are missing",
			pathID:         validRestUUID,
			body:           `{"name":""}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 404 when restaurant not found",
			pathID: validRestUUID,
			body:   `{"name":"Menu Principal"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("CreateMenu", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, common.NewNotFoundErr(errors.New("restaurant not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "should return 500 when service fails with unexpected error",
			pathID: validRestUUID,
			body:   `{"name":"Menu Principal"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("CreateMenu", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			req := httptest.NewRequest(http.MethodPost, "/restaurant/"+tt.pathID+"/menu", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_GetActiveMenu(t *testing.T) {
	validRestUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockCatalogQueryService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should return active menu and 200",
			pathID: validRestUUID,
			mockSetup: func(svc *mockCatalogQueryService) {
				svc.On("GetActiveMenu", mock.Anything, mock.Anything).
					Return(&input.MenuResponse{
						ID: "menu-uuid", Name: "Menu Ativo", RestaurantID: validRestUUID, Status: "ACTIVE",
						Categories: []*input.CategoryResponse{},
					}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"name":"Menu Ativo"`,
		},
		{
			name:           "should return error when restaurant ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockCatalogQueryService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 404 when no active menu found",
			pathID: validRestUUID,
			mockSetup: func(svc *mockCatalogQueryService) {
				svc.On("GetActiveMenu", mock.Anything, mock.Anything).
					Return(nil, common.NewNotFoundErr(errors.New("no active menu"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(qSvc)
			mux := setupMenuHandler(svc, qSvc)

			req := httptest.NewRequest(http.MethodGet, "/restaurant/"+tt.pathID+"/menu", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_ActivateMenu(t *testing.T) {
	validMenuUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should activate menu and return 200",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ActiveMenu", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"menu activated"`,
		},
		{
			name:           "should return error when menu ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 422 when menu is already active",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ActiveMenu", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrAlreadyActive)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"menu is already active"`,
		},
		{
			name:   "should return 422 when menu has no items",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ActiveMenu", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrCannotActivateEmpty)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"cannot active an empty menu"`,
		},
		{
			name:   "should return 404 when menu not found",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ActiveMenu", mock.Anything, mock.Anything).
					Return(common.NewNotFoundErr(errors.New("menu not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			req := httptest.NewRequest(http.MethodPatch, "/menu/"+tt.pathID+"/activate", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_ArchiveMenu(t *testing.T) {
	validMenuUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should archive menu and return 200",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ArchiveMenu", mock.Anything, mock.Anything).
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"menu archived"`,
		},
		{
			name:           "should return error when menu ID is not a valid UUID",
			pathID:         "not-a-uuid",
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "should return 422 when menu is already archived",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ArchiveMenu", mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrAlreadyArchived)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `"error":"menu is already archived"`,
		},
		{
			name:   "should return 404 when menu not found",
			pathID: validMenuUUID,
			mockSetup: func(svc *mockMenuService) {
				svc.On("ArchiveMenu", mock.Anything, mock.Anything).
					Return(common.NewNotFoundErr(errors.New("menu not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			req := httptest.NewRequest(http.MethodPatch, "/menu/"+tt.pathID+"/archive", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_AddCategory(t *testing.T) {
	validMenuUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathMenuID     string
		body           string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "should add category and return 201",
			pathMenuID: validMenuUUID,
			body:       `{"name":"Pizzas"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddCategory", mock.Anything, mock.Anything, mock.AnythingOfType("AddCategoryRequest")).
					Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"message":"a new category has been added to menu"`,
		},
		{
			name:           "should return error when menu ID is not a valid UUID",
			pathMenuID:     "not-a-uuid",
			body:           `{"name":"Pizzas"}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when body is malformed JSON",
			pathMenuID:     validMenuUUID,
			body:           `{invalid`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when required fields are missing",
			pathMenuID:     validMenuUUID,
			body:           `{"name":""}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 422 when menu is not in draft status",
			pathMenuID: validMenuUUID,
			body:       `{"name":"Pizzas"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddCategory", mock.Anything, mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrMenuNotEditable)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "should return 422 when max categories reached",
			pathMenuID: validMenuUUID,
			body:       `{"name":"Pizzas"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddCategory", mock.Anything, mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrMaxCategoryReached)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			req := httptest.NewRequest(http.MethodPost, "/menu/"+tt.pathMenuID+"/categories", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_AddItemToCategory(t *testing.T) {
	validMenuUUID := uuid.New().String()
	validCatUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathMenuID     string
		pathCatID      string
		body           string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "should add item to category and return 201",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			body:       `{"name":"Margherita","description":"Classic pizza","price_cents":2500}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddItemToCategory", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("AddItemRequest")).
					Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"message":"a new item has been added to category"`,
		},
		{
			name:           "should return error when menu ID is not a valid UUID",
			pathMenuID:     "not-a-uuid",
			pathCatID:      validCatUUID,
			body:           `{"name":"Margherita","description":"Classic pizza","price_cents":2500}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return error when category ID is not a valid UUID",
			pathMenuID:     validMenuUUID,
			pathCatID:      "not-a-uuid",
			body:           `{"name":"Margherita","description":"Classic pizza","price_cents":2500}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when body is malformed JSON",
			pathMenuID:     validMenuUUID,
			pathCatID:      validCatUUID,
			body:           `{invalid`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when required fields are missing",
			pathMenuID:     validMenuUUID,
			pathCatID:      validCatUUID,
			body:           `{"name":""}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 422 when menu is not in draft status",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			body:       `{"name":"Margherita","description":"Classic pizza","price_cents":2500}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddItemToCategory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(common.NewBusinessRuleErr(menu.ErrMenuNotEditable)).Once()
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "should return 404 when category not found in menu",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			body:       `{"name":"Margherita","description":"Classic pizza","price_cents":2500}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("AddItemToCategory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(common.NewNotFoundErr(menu.ErrCategoryNotFound)).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			path := "/menu/" + tt.pathMenuID + "/categories/" + tt.pathCatID + "/item"
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}

func TestMenuHandler_UpdateItem(t *testing.T) {
	validMenuUUID := uuid.New().String()
	validCatUUID := uuid.New().String()
	validItemUUID := uuid.New().String()

	tests := []struct {
		name           string
		pathMenuID     string
		pathCatID      string
		pathItemID     string
		body           string
		mockSetup      func(svc *mockMenuService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "should update item and return 200",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			pathItemID: validItemUUID,
			body:       `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("UpdateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("UpdateItemRequest")).
					Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"item updated"`,
		},
		{
			name:           "should return error when menu ID is not a valid UUID",
			pathMenuID:     "not-a-uuid",
			pathCatID:      validCatUUID,
			pathItemID:     validItemUUID,
			body:           `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return error when category ID is not a valid UUID",
			pathMenuID:     validMenuUUID,
			pathCatID:      "not-a-uuid",
			pathItemID:     validItemUUID,
			body:           `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return error when item ID is not a valid UUID",
			pathMenuID:     validMenuUUID,
			pathCatID:      validCatUUID,
			pathItemID:     "not-a-uuid",
			body:           `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when body is malformed JSON",
			pathMenuID:     validMenuUUID,
			pathCatID:      validCatUUID,
			pathItemID:     validItemUUID,
			body:           `{invalid`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should return 400 when required fields are missing",
			pathMenuID:     validMenuUUID,
			pathCatID:      validCatUUID,
			pathItemID:     validItemUUID,
			body:           `{}`,
			mockSetup:      func(svc *mockMenuService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "should return 404 when category not found",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			pathItemID: validItemUUID,
			body:       `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("UpdateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(common.NewNotFoundErr(menu.ErrCategoryNotFound)).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "should return 500 when service fails with unexpected error",
			pathMenuID: validMenuUUID,
			pathCatID:  validCatUUID,
			pathItemID: validItemUUID,
			body:       `{"price_cents":3000,"status":"AVAILABLE"}`,
			mockSetup: func(svc *mockMenuService) {
				svc.On("UpdateItem", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockMenuService)
			qSvc := new(mockCatalogQueryService)
			tt.mockSetup(svc)
			mux := setupMenuHandler(svc, qSvc)

			path := "/menu/" + tt.pathMenuID + "/categories/" + tt.pathCatID + "/items/" + tt.pathItemID
			req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			svc.AssertExpectations(t)
			qSvc.AssertExpectations(t)
		})
	}
}
