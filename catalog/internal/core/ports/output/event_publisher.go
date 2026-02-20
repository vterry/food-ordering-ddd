package output

import "context"

type EventMessage struct {
	EventID       string
	AggregateID   string
	AggregateType string
	EventType     string
	OcurredAt     int64
	Payload       []byte
}

type EventPublisher interface {
	Publish(ctx context.Context, msg EventMessage) error
}
