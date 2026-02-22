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

type UnrecoverableErr interface {
	error
	IsUnrecoverable()
}

type InfraConnectionErr interface {
	error
	IsInfraConnection()
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

type unrecoverableErr struct{ cause error }

func (e *unrecoverableErr) Error() string    { return e.cause.Error() }
func (e *unrecoverableErr) Unwrap() error    { return e.cause }
func (e *unrecoverableErr) IsUnrecoverable() {}

func NewUnrecoverableErr(cause error) error {
	return &unrecoverableErr{
		cause: cause,
	}
}

type infraConnectionErr struct{ cause error }

func (e *infraConnectionErr) Error() string      { return e.cause.Error() }
func (e *infraConnectionErr) Unwrap() error      { return e.cause }
func (e *infraConnectionErr) IsInfraConnection() {}

func NewInfraConnectionErr(cause error) error {
	return &infraConnectionErr{
		cause: cause,
	}
}
