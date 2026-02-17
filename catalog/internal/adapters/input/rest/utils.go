package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

var (
	Validate = validator.New()
)

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func handleInputValidation(w http.ResponseWriter, req *http.Request, payload any) error {

	if err := ParseJSON(req, payload); err != nil {
		return err
	}

	if err := Validate.Struct(payload); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return err
		}
		return err
	}
	return nil
}

// TODO - Melhorar esse tratamento
func handleAppError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	var validationErrors validator.ValidationErrors
	var syntaxError *json.SyntaxError

	switch {
	case errors.As(err, &validationErrors):
		statusCode = http.StatusBadRequest
	case errors.Is(err, output.ErrEntityNotFound):
		statusCode = http.StatusNotFound
	case errors.As(err, &syntaxError):
		statusCode = http.StatusBadRequest
	case errors.Is(err, menu.ErrMenuNotEditable),
		errors.Is(err, menu.ErrCannotActivateEmpty),
		errors.Is(err, menu.ErrAlreadyActive),
		errors.Is(err, menu.ErrAlreadyArchived),
		errors.Is(err, restaurant.ErrAlreadyOpened),
		errors.Is(err, restaurant.ErrAlreadyClosed),
		errors.Is(err, restaurant.ErrNoActiveMenu):
		statusCode = http.StatusUnprocessableEntity
	case errors.Is(err, menu.ErrCategoryNotFound):
		statusCode = http.StatusNotFound
	}

	if statusCode == http.StatusInternalServerError {
		slog.Error("internal error", "err", err)
		WriteJSON(w, statusCode, map[string]string{"error": "internal error"})
		return
	}

	WriteJSON(w, statusCode, map[string]string{"error": err.Error()})
}
