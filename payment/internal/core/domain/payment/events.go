package payment

import "time"

// PaymentAuthorized is triggered when a payment is successfully authorized by the gateway.
type PaymentAuthorized struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentAuthorizationFailed is triggered when the gateway rejects the authorization.
type PaymentAuthorizationFailed struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentCaptured is triggered when funds are successfully captured.
type PaymentCaptured struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentCaptureFailed is triggered when capture fails.
type PaymentCaptureFailed struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentReleased is triggered when an authorized payment is released (voided).
type PaymentReleased struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentRefunded is triggered when a captured payment is refunded.
type PaymentRefunded struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
}
