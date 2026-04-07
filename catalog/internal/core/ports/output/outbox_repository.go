package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

const DefaultClaimTTL = 5 * time.Minute

type OutboxEvent struct {
	UUID          uuid.UUID
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       []byte
	OccurredOn    time.Time
	RetryCount    int
	ClaimedBy     *string
	ClaimedAt     *time.Time
}

type OutboxRepository interface {
	SaveEvents(ctx context.Context, aggregateID, aggregateType string, events []common.DomainEvent) error
	ClaimAndFindEvents(ctx context.Context, processorID string, limit int, claimTTL time.Duration) ([]OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventID uuid.UUID) error
	IncrementRetry(ctx context.Context, eventID uuid.UUID) error
	MoveToDLQ(ctx context.Context, event OutboxEvent, reason string) error
}
