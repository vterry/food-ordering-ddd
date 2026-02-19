package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

// httpStatusFor maps domain and infrastructure errors to HTTP status codes.
// Unknown errors (infrastructure/unexpected) map to 500.
func httpStatusFor(err error) int {
	var validationErrors validator.ValidationErrors
	var syntaxError *json.SyntaxError

	switch {
	case errors.As(err, &validationErrors):
		return http.StatusBadRequest
	case errors.As(err, &syntaxError):
		return http.StatusBadRequest
	case errors.Is(err, output.ErrEntityNotFound):
		return http.StatusNotFound
	case errors.Is(err, menu.ErrCategoryNotFound):
		return http.StatusNotFound
	case errors.Is(err, menu.ErrMenuNotEditable),
		errors.Is(err, menu.ErrCannotActivateEmpty),
		errors.Is(err, menu.ErrAlreadyActive),
		errors.Is(err, menu.ErrAlreadyArchived),
		errors.Is(err, restaurant.ErrAlreadyOpened),
		errors.Is(err, restaurant.ErrAlreadyClosed),
		errors.Is(err, restaurant.ErrNoActiveMenu):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
