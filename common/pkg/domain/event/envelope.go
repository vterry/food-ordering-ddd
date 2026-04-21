package event

import (
	"time"
	"github.com/vterry/food-project/common/pkg/domain/base"
)

// EventEnvelope wraps a domain event with cross-cutting metadata.
type EventEnvelope struct {
	Header  Metadata    `json:"header"`
	Payload interface{} `json:"payload"`
}

// Metadata contains traceability and origin information.
type Metadata struct {
	ID            string    `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"`
	Source        string    `json:"source"`
}

// Wrap converts a domain event into an envelope.
func Wrap(event base.DomainEvent, correlationID string, source string) EventEnvelope {
	return EventEnvelope{
		Header: Metadata{
			ID:            event.EventID().String(),
			CorrelationID: correlationID,
			Timestamp:     event.OccurredAt(),
			Type:          event.EventType(),
			Source:        source,
		},
		Payload: event,
	}
}
