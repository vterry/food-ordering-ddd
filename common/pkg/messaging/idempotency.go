package messaging

import (
	"context"
	"fmt"
	"log/slog"
)

type IdempotencyRepository interface {
	IsMessageProcessed(ctx context.Context, messageID string) (bool, error)
	MarkMessageAsProcessed(ctx context.Context, messageID string) error
}

type IdempotentHandler struct {
	repo IdempotencyRepository
}

func NewIdempotentHandler(repo IdempotencyRepository) *IdempotentHandler {
	return &IdempotentHandler{repo: repo}
}

func (h *IdempotentHandler) Handle(ctx context.Context, messageID string, handler func(ctx context.Context) error) error {
	if messageID == "" {
		slog.Warn("Processing message without ID, idempotency cannot be guaranteed")
		return handler(ctx)
	}

	processed, err := h.repo.IsMessageProcessed(ctx, messageID)
	if err != nil {
		return fmt.Errorf("checking idempotency: %w", err)
	}

	if processed {
		slog.Info("Message already processed, skipping", "message_id", messageID)
		return nil
	}

	if err := handler(ctx); err != nil {
		return err
	}

	if err := h.repo.MarkMessageAsProcessed(ctx, messageID); err != nil {
		slog.Error("Failed to mark message as processed", "message_id", messageID, "error", err)
		// We don't return error here because the handler already succeeded.
		// If this fails, the next retry will re-run the handler.
		// In a transactional repo, this should be done within the same TX.
	}

	return nil
}
