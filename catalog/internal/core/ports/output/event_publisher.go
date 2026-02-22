package output

import (
	"context"
	"encoding/json"
)

type EventMessage struct {
	EventID       string
	AggregateID   string
	AggregateType string
	EventType     string
	OcurredAt     int64
	Payload       json.RawMessage
}

type EventPublisher interface {
	Publish(ctx context.Context, msg EventMessage) error
}
