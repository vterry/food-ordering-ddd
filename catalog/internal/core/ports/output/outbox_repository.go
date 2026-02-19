package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
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
	SaveEvents(ctx context.Context, aggregateID, aggregateType string, events []common.DomainEvent) error
	FindUnpublishedEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventID uuid.UUID) error
}
