package payment

import (
	"time"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

// PaymentAuthorized is triggered when a payment is successfully authorized by the gateway.
type PaymentAuthorized struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentAuthorized) EventID() vo.ID { return vo.NewID("ev-pay-auth-" + e.PaymentID) }
func (e PaymentAuthorized) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentAuthorized) EventType() string { return "payment.authorized" }

// PaymentAuthorizationFailed is triggered when the gateway rejects the authorization.
type PaymentAuthorizationFailed struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Reason     string    `json:"reason"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentAuthorizationFailed) EventID() vo.ID { return vo.NewID("ev-pay-auth-fail-" + e.PaymentID) }
func (e PaymentAuthorizationFailed) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentAuthorizationFailed) EventType() string { return "payment.authorization_failed" }

// PaymentCaptured is triggered when funds are successfully captured.
type PaymentCaptured struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentCaptured) EventID() vo.ID { return vo.NewID("ev-pay-cap-" + e.PaymentID) }
func (e PaymentCaptured) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentCaptured) EventType() string { return "payment.captured" }

// PaymentCaptureFailed is triggered when capture fails.
type PaymentCaptureFailed struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Reason     string    `json:"reason"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentCaptureFailed) EventID() vo.ID { return vo.NewID("ev-pay-cap-fail-" + e.PaymentID) }
func (e PaymentCaptureFailed) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentCaptureFailed) EventType() string { return "payment.capture_failed" }

// PaymentReleased is triggered when an authorized payment is released (voided).
type PaymentReleased struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentReleased) EventID() vo.ID { return vo.NewID("ev-pay-rel-" + e.PaymentID) }
func (e PaymentReleased) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentReleased) EventType() string { return "payment.released" }

// PaymentRefunded is triggered when a captured payment is refunded.
type PaymentRefunded struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	OccurredAtTime time.Time `json:"occurred_at"`
}

func (e PaymentRefunded) EventID() vo.ID { return vo.NewID("ev-pay-ref-" + e.PaymentID) }
func (e PaymentRefunded) OccurredAt() time.Time { return e.OccurredAtTime }
func (e PaymentRefunded) EventType() string { return "payment.refunded" }
