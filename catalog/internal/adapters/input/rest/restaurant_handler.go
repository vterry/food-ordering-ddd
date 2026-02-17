package rest

import (
	"log/slog"
	"net/http"

	"github.com/vterry/food-ordering/catalog/internal/adapters/input/rest/middleware"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

type RestaurantHandler struct {
	appService input.RestaurantService
	logger     *slog.Logger
}

func NewRestaurantHandler(appService input.RestaurantService, logger *slog.Logger) *RestaurantHandler {
	return &RestaurantHandler{
		appService: appService,
		logger:     logger,
	}
}

func (r *RestaurantHandler) RegisterRoutes(mux *http.ServeMux) {
	loggerMw := middleware.LoggerMiddleware(r.logger)

	chain := func(h http.HandlerFunc) http.Handler {
		return middleware.Chain(h, loggerMw)
	}

	mux.Handle("POST /restaurant", chain(r.handleCreateRestaurant))
	mux.Handle("GET /restaurant/{id}", chain(r.handleGetRestaurant))
	mux.Handle("PATCH /restaurant/{id}/open", chain(r.handleOpenRestaurant))
	mux.Handle("PATCH /restaurant/{id}/close", chain(r.handleCloseRestaurant))
}

func (r *RestaurantHandler) handleCreateRestaurant(res http.ResponseWriter, req *http.Request) {
	var payload input.CreateRestaurantRequest

	if err := handleInputValidation(res, req, &payload); err != nil {
		handleAppError(res, err)
		return
	}

	restaurantResp, err := r.appService.CreateRestaurant(req.Context(), payload)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusCreated, restaurantResp); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

}

func (r *RestaurantHandler) handleGetRestaurant(res http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	restaurantResp, err := r.appService.GetRestaurant(req.Context(), idStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, restaurantResp); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

}

func (r *RestaurantHandler) handleOpenRestaurant(res http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")

	err := r.appService.OpenRestaurant(req.Context(), idStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "restaurant opened"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (r *RestaurantHandler) handleCloseRestaurant(res http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")

	err := r.appService.CloseRestaurant(req.Context(), idStr)
	if err != nil {
		handleAppError(res, err)
		return
	}

	if err := WriteJSON(res, http.StatusOK, map[string]string{"message": "restaurant closed"}); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
