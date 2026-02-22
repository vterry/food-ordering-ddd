package common

type NotFoundErr interface {
	error
	IsNotFound()
}

type BusinessRuleErr interface {
	error
	IsBusinessRule()
}

type ValidationErr interface {
	error
	IsValidation()
}

type notFoundError struct{ cause error }

func (e *notFoundError) Error() string { return e.cause.Error() }
func (e *notFoundError) Unwrap() error { return e.cause }
func (e *notFoundError) IsNotFound()   {}

func NewNotFoundErr(cause error) error {
	return &notFoundError{cause: cause}
}

type businessRuleError struct{ cause error }

func (e *businessRuleError) Error() string   { return e.cause.Error() }
func (e *businessRuleError) Unwrap() error   { return e.cause }
func (e *businessRuleError) IsBusinessRule() {}

func NewBusinessRuleErr(cause error) error {
	return &businessRuleError{
		cause: cause,
	}
}

type validationErr struct{ cause error }

func (e *validationErr) Error() string { return e.cause.Error() }
func (e *validationErr) Unwrap() error { return e.cause }
func (e *validationErr) IsValidation() {}

func NewValidationErr(cause error) error {
	return &validationErr{
		cause: cause,
	}
}
