package payment

import "fmt"

// Status represents the state of a payment in its lifecycle.
type Status string

const (
	StatusCreated             Status = "CREATED"
	StatusAuthorized          Status = "AUTHORIZED"
	StatusCaptured            Status = "CAPTURED"
	StatusAuthorizationFailed Status = "AUTHORIZATION_FAILED"
	StatusReleased            Status = "RELEASED"
	StatusRefunded            Status = "REFUNDED"
)

// IsValid checks if the status is one of the predefined valid statuses.
func (s Status) IsValid() error {
	switch s {
	case StatusCreated, StatusAuthorized, StatusCaptured, StatusAuthorizationFailed, StatusReleased, StatusRefunded:
		return nil
	default:
		return fmt.Errorf("invalid payment status: %s", s)
	}
}
