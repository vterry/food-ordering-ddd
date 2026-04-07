package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository/dao"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

var _ output.OutboxRepository = (*OutboxRepository)(nil)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (o *OutboxRepository) SaveEvents(ctx context.Context, aggregateID, aggregateType string, events []common.DomainEvent) error {
	executor := getExecutor(ctx, o.db)

	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.EventName(), err)
		}

		_, err = executor.ExecContext(ctx, QueryInsertOutboxEvent,
			event.EventID().String(),
			aggregateID,
			aggregateType,
			event.EventName(),
			payload,
			event.OccurredOn(),
		)
		if err != nil {
			return fmt.Errorf("failed to save outbox event %s: %w", event.EventName(), err)
		}
	}

	return nil
}

func (o *OutboxRepository) ClaimAndFindEvents(ctx context.Context, processorID string, limit int, claimTTL time.Duration) ([]output.OutboxEvent, error) {
	executor := getExecutor(ctx, o.db)

	claimTTLSeconds := int(claimTTL.Seconds())
	_, err := executor.ExecContext(ctx, QueryClaimOutboxEvents, processorID, claimTTLSeconds, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to claim outbox events: %w", err)
	}

	rows, err := executor.QueryContext(ctx, QueryFindClaimedEvents, processorID)
	if err != nil {
		return nil, fmt.Errorf("failed to find claimed events: %w", err)
	}
	defer rows.Close()

	events := make([]output.OutboxEvent, 0)
	for rows.Next() {
		var eventDao dao.OutboxEventDAO
		err := rows.Scan(
			&eventDao.ID,
			&eventDao.UUID,
			&eventDao.AggregateID,
			&eventDao.AggregateType,
			&eventDao.EventType,
			&eventDao.Payload,
			&eventDao.OccurredOn,
			&eventDao.RetryCount,
		)

		if err != nil {
			return nil, err
		}

		eventID, err := uuid.Parse(eventDao.UUID)
		if err != nil {
			return nil, err
		}

		events = append(events, output.OutboxEvent{
			UUID:          eventID,
			AggregateID:   eventDao.AggregateID,
			AggregateType: eventDao.AggregateType,
			EventType:     eventDao.EventType,
			Payload:       eventDao.Payload,
			OccurredOn:    eventDao.OccurredOn,
			RetryCount:    eventDao.RetryCount,
		})
	}

	return events, nil
}

func (o *OutboxRepository) MarkAsPublished(ctx context.Context, eventID uuid.UUID) error {
	executor := getExecutor(ctx, o.db)
	_, err := executor.ExecContext(ctx, QueryMarkAsPublished, eventID.String())
	return err
}

func (o *OutboxRepository) IncrementRetry(ctx context.Context, eventID uuid.UUID) error {
	_, err := getExecutor(ctx, o.db).ExecContext(ctx, QueryIncrementRetryCount, eventID.String())
	return err
}

func (o *OutboxRepository) MoveToDLQ(ctx context.Context, event output.OutboxEvent, reason string) error {
	executor := getExecutor(ctx, o.db)
	_, err := executor.ExecContext(ctx, QueryInsertOutboxDLQ, event.UUID.String(), event.AggregateID, event.AggregateType, event.EventType, event.Payload, event.OccurredOn, event.RetryCount, reason)
	if err != nil {
		return err
	}

	_, err = executor.ExecContext(ctx, QueryDeleteFromOutbox, event.UUID.String())
	return err
}
