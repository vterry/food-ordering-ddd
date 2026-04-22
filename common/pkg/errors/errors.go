package errors

import (
	"fmt"
	"net/http"
)

// ErrorType defines the category of the error.
type ErrorType string

const (
	ErrorTypeDomain         ErrorType = "DOMAIN"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"
	ErrorTypeInfrastructure ErrorType = "INFRASTRUCTURE"
	ErrorTypeInternal       ErrorType = "INTERNAL"
)

// AppError is the standardized error structure for the platform.
type AppError struct {
	Slug    string    `json:"slug"`
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
	Status  int       `json:"-"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s (Slug: %s): %v", e.Type, e.Message, e.Slug, e.Err)
	}
	return fmt.Sprintf("[%s] %s (Slug: %s)", e.Type, e.Message, e.Slug)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Helper functions for creating errors

func NewDomainError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusBadRequest, Type: ErrorTypeDomain, Err: err}
}

func NewNotFoundError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusNotFound, Type: ErrorTypeNotFound, Err: err}
}

func NewConflictError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusConflict, Type: ErrorTypeConflict, Err: err}
}

func NewUnauthorizedError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusUnauthorized, Type: ErrorTypeUnauthorized, Err: err}
}

func NewInfrastructureError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusInternalServerError, Type: ErrorTypeInfrastructure, Err: err}
}

func NewInternalError(slug, message string, err error) *AppError {
	return &AppError{Slug: slug, Message: message, Status: http.StatusInternalServerError, Type: ErrorTypeInternal, Err: err}
}
