package output

import "errors"

var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrValidation     = errors.New("validation failed")
)
