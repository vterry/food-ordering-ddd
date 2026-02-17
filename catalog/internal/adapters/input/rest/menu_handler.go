package rest

import (
	"log/slog"
	"net/http"

	"github.com/vterry/food-ordering/catalog/internal/adapters/input/rest/middleware"
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
	mux.Handle("POST /menu/{menu_id}/categories", chain(m.handleAddMenuCategorie))
	mux.Handle("POST /menu/{menu_id}/categories/{category_id}/item", chain(m.handleAddItemToCategory))
	mux.Handle("PUT /menu/{menu_id}/categories/{category_id}/items/{item_id}", chain(m.handleUpdateItem))

}

func (m *MenuHandler) handleCreateMenu(res http.ResponseWriter, req *http.Request) {
	restauntIdStr := req.PathValue("id")

	var payload input.CreateMenuRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, err)
		return
	}

	menuResponse, err := m.appService.CreateMenu(req.Context(), restauntIdStr, payload)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, menuResponse); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleGetActiveMenu(res http.ResponseWriter, req *http.Request) {
	restauntIdStr := req.PathValue("id")

	menuResponse, err := m.appService.GetActiveMenu(req.Context(), restauntIdStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, menuResponse); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleActivateMenu(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("id")

	err := m.appService.ActiveMenu(req.Context(), menuIdStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "menu activated"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleArchiveMenu(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("id")

	err := m.appService.ArchiveMenu(req.Context(), menuIdStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "menu archived"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleAddMenuCategorie(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")

	var payload input.AddCategoryRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, err)
		return
	}

	err := m.appService.AddCategory(req.Context(), menuIdStr, payload)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, map[string]string{"message": "a new category has been added to menu"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleAddItemToCategory(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")
	categoryIdStr := req.PathValue("category_id")

	var payload input.AddItemRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, err)
		return
	}

	err := m.appService.AddItemToCategory(req.Context(), menuIdStr, categoryIdStr, payload)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, map[string]string{"message": "a new item has been added to category"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (m *MenuHandler) handleUpdateItem(res http.ResponseWriter, req *http.Request) {
	menuIdStr := req.PathValue("menu_id")
	categoryIdStr := req.PathValue("category_id")
	itemIdStr := req.PathValue("item_id")

	var payload input.UpdateItemRequest
	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, err)
		return
	}

	err := m.appService.UpdateItem(req.Context(), menuIdStr, categoryIdStr, itemIdStr, payload)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "item updated"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
