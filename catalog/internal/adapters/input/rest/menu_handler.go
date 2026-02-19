package rest

import (
	"log/slog"
	"net/http"

	"github.com/vterry/food-ordering/catalog/internal/adapters/input/rest/middleware"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

type MenuHandler struct {
	appService input.MenuService
	logger     *slog.Logger
}

func NewMenuHandler(appService input.MenuService, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{
		appService: appService,
		logger:     logger,
	}
}

func (m *MenuHandler) RegisterRoutes(mux *http.ServeMux) {
	loggerMw := middleware.LoggerMiddleware(m.logger)

	chain := func(h http.HandlerFunc) http.Handler {
		return middleware.Chain(h, loggerMw)
	}

	mux.Handle("POST /restaurant/{id}/menu", chain(m.handleCreateMenu))
	mux.Handle("GET /restaurant/{id}/menu", chain(m.handleGetActiveMenu))
	mux.Handle("PATCH /menu/{id}/activate", chain(m.handleActivateMenu))
	mux.Handle("PATCH /menu/{id}/archive", chain(m.handleArchiveMenu))
	mux.Handle("POST /menu/{menu_id}/categories", chain(m.handleAddCategory))
	mux.Handle("POST /menu/{menu_id}/categories/{category_id}/item", chain(m.handleAddItemToCategory))
	mux.Handle("PUT /menu/{menu_id}/categories/{category_id}/items/{item_id}", chain(m.handleUpdateItem))

}

func (m *MenuHandler) handleCreateMenu(res http.ResponseWriter, req *http.Request) {
	restaurantIdStr := req.PathValue("id")
	restaurantId, err := valueobjects.ParseRestaurantId(restaurantIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	var payload input.CreateMenuRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, req, err)
		return
	}

	menuResponse, err := m.appService.CreateMenu(req.Context(), restaurantId, payload)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, menuResponse); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleGetActiveMenu(res http.ResponseWriter, req *http.Request) {
	restaurantIdStr := req.PathValue("id")
	restaurantId, err := valueobjects.ParseRestaurantId(restaurantIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	menuResponse, err := m.appService.GetActiveMenu(req.Context(), restaurantId)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, menuResponse); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleActivateMenu(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("id")
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	err = m.appService.ActiveMenu(req.Context(), menuId)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "menu activated"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleArchiveMenu(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("id")
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	err = m.appService.ArchiveMenu(req.Context(), menuId)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "menu archived"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleAddCategory(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	var payload input.AddCategoryRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, req, err)
		return
	}

	err = m.appService.AddCategory(req.Context(), menuId, payload)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, map[string]string{"message": "a new category has been added to menu"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleAddItemToCategory(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	categoryIdStr := req.PathValue("category_id")
	categoryId, err := valueobjects.ParseCategoryId(categoryIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	var payload input.AddItemRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, req, err)
		return
	}

	err = m.appService.AddItemToCategory(req.Context(), menuId, categoryId, payload)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, map[string]string{"message": "a new item has been added to category"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleUpdateItem(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	categoryIdStr := req.PathValue("category_id")
	categoryId, err := valueobjects.ParseCategoryId(categoryIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	itemIdStr := req.PathValue("item_id")
	itemId, err := valueobjects.ParseItemId(itemIdStr)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	var payload input.UpdateItemRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, req, err)
		return
	}

	err = m.appService.UpdateItem(req.Context(), menuId, categoryId, itemId, payload)
	if err != nil {
		handleAppError(res, req, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "item updated"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
