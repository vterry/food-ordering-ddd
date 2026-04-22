package errors

import (
	"errors"
	"net/http"
	"strings"
)

// HttpErrorResponse is the standardized error response format for HTTP.
type HttpErrorResponse struct {
	Message string            `json:"message"`
	Slug    string            `json:"slug"`
	Details []HttpErrorDetail `json:"details,omitempty"`
}

// HttpErrorDetail provides additional context for an error in HTTP responses.
type HttpErrorDetail struct {
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	ErrorSlug  string `json:"error_slug,omitempty"`
	Message    string `json:"message,omitempty"`
}

// MapToHTTP converts a generic error (or AppError) into a structured HTTP response and status code.
func MapToHTTP(err error) (HttpErrorResponse, int) {
	publicError := "Internal Server Error"
	statusCode := http.StatusInternalServerError

	errorSlug := strings.ToLower(strings.ReplaceAll(publicError, " ", "_"))

	var appErr *AppError
	if errors.As(err, &appErr) {
		publicError = appErr.Message
		errorSlug = strings.ToLower(appErr.Slug)

		if appErr.Status != 0 {
			statusCode = appErr.Status
		}
	}

	return HttpErrorResponse{
		Slug:    errorSlug,
		Message: publicError,
	}, statusCode
}
