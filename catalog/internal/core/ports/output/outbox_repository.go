package output

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	UUID          uuid.UUID
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	OccurredOn    time.Time
}

type OutboxRepository interface {
	FindUnpublishedEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventID uuid.UUID) error
}
