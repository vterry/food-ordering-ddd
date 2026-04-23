package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type OutboxMessage struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	CorrelationID string
}

type Repository interface {
	FetchUnpublished(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkAsPublished(ctx context.Context, id string) error
}

type Publisher interface {
	PublishRaw(ctx context.Context, eventType string, correlationID string, payload []byte) error
}

type Relay struct {
	repo     Repository
	pub      Publisher
	interval time.Duration
	limit    int
}

func NewRelay(repo Repository, pub Publisher, interval time.Duration, limit int) *Relay {
	if interval == 0 {
		interval = 500 * time.Millisecond
	}
	if limit == 0 {
		limit = 50
	}
	return &Relay{
		repo:     repo,
		pub:      pub,
		interval: interval,
		limit:    limit,
	}
}

func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	slog.Info("Outbox Relay started", "interval", r.interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox Relay stopping...")
			return
		case <-ticker.C:
			if err := r.processMessages(ctx); err != nil {
				slog.Error("Failed to process outbox messages", "error", err)
			}
		}
	}
}

func (r *Relay) processMessages(ctx context.Context) error {
	messages, err := r.repo.FetchUnpublished(ctx, r.limit)
	if err != nil {
		return fmt.Errorf("fetching unpublished messages: %w", err)
	}

	if len(messages) == 0 {
		return nil
	}

	slog.Debug("Processing outbox messages", "count", len(messages))

	for _, msg := range messages {
		err := r.pub.PublishRaw(ctx, msg.EventType, msg.CorrelationID, msg.Payload)
		if err != nil {
			slog.Error("Failed to publish outbox message", "id", msg.ID, "error", err)
			continue
		}

		if err := r.repo.MarkAsPublished(ctx, msg.ID); err != nil {
			slog.Error("Failed to mark outbox message as published", "id", msg.ID, "error", err)
			// We might publish twice if this fails, but it's better than losing the message.
			// The consumer should handle idempotency.
		}
	}

	return nil
}
