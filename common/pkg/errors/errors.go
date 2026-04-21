package errors

import (
	"fmt"
)

// ErrorType defines the category of the error.
type ErrorType string

const (
	ErrorTypeDomain         ErrorType = "DOMAIN"         // Business rule violation
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"      // Resource not found
	ErrorTypeConflict       ErrorType = "CONFLICT"       // State conflict (e.g. duplicate)
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"   // Auth issues
	ErrorTypeInfrastructure ErrorType = "INFRASTRUCTURE" // External systems failure
	ErrorTypeInternal       ErrorType = "INTERNAL"       // Unexpected failure
)

// AppError is the standardized error structure for the platform.
type AppError struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Type    ErrorType `json:"type"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Helper functions for creating errors

func NewDomainError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Type: ErrorTypeDomain, Err: err}
}

func NewNotFoundError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Type: ErrorTypeNotFound, Err: err}
}

func NewConflictError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Type: ErrorTypeConflict, Err: err}
}

func NewInfrastructureError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Type: ErrorTypeInfrastructure, Err: err}
}

func NewInternalError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Type: ErrorTypeInternal, Err: err}
}
