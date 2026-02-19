package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator"
)

var validate = validator.New()

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
	if err := validate.Struct(payload); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return err
		}
		return err
	}
	return nil
}

func handleAppError(w http.ResponseWriter, req *http.Request, err error) {
	code := httpStatusFor(err)
	if code == http.StatusInternalServerError {
		slog.Error("internal error", "err", err, "method", req.Method, "path", req.URL.Path)
		WriteJSON(w, code, map[string]string{"error": "internal error"})
		return
	}
	WriteJSON(w, code, map[string]string{"error": err.Error()})
}
