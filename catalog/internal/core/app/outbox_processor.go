package app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

type OutboxProcessor struct {
	outboxRepo   output.OutboxRepository
	publisher    output.EventPublisher
	uow          output.UnitOfWork
	pollInterval time.Duration
	batchSize    int
	retryCount   int
	logger       *slog.Logger
}

func NewOutboxProcessor(repo output.OutboxRepository, pub output.EventPublisher, uow output.UnitOfWork, pollInterval time.Duration, batchSize int, retryCount int, logger *slog.Logger) *OutboxProcessor {

	return &OutboxProcessor{
		outboxRepo:   repo,
		publisher:    pub,
		uow:          uow,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		retryCount:   retryCount,
		logger:       logger,
	}
}

func (o *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(o.pollInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				o.logger.Info("OutboxProcessor stopped due to context cancellation")
				return
			case <-ticker.C:
				if err := o.processBatch(ctx); err != nil {
					o.logger.Error("OutboxProcessor errored processing batch", "error", err)
				}
			}
		}
	}()
}

func (o *OutboxProcessor) processBatch(ctx context.Context) error {
	return o.uow.Run(ctx, func(txCtx context.Context) error {
		eventsToProcess, err := o.outboxRepo.FindUnpublishedEvents(txCtx, o.batchSize)
		if err != nil {
			return err
		}

		if len(eventsToProcess) == 0 {
			return nil
		}

		for _, event := range eventsToProcess {
			msg := output.EventMessage{
				EventID:       event.UUID.String(),
				AggregateID:   event.AggregateID,
				AggregateType: event.AggregateType,
				EventType:     event.EventType,
				Payload:       json.RawMessage(event.Payload),
				OcurredAt:     event.OccurredOn.UnixMilli(),
			}

			pubErr := o.publisher.Publish(ctx, msg)

			if pubErr != nil {
				var unrecoverableErr common.UnrecoverableErr
				if errors.As(pubErr, &unrecoverableErr) {
					if mErr := o.outboxRepo.MoveToDLQ(txCtx, event, pubErr.Error()); mErr != nil {
						o.logger.Error("failed to move unrecoverable event to DLQ", "event_id", event.UUID, "error", mErr)
						return mErr
					}
					continue
				}

				var infraErr common.InfraConnectionErr
				if errors.As(pubErr, &infraErr) {
					o.logger.Error("broker connection failure. aborting batch processing", "error", pubErr)
					return pubErr
				}

				if event.RetryCount >= o.retryCount {
					o.logger.Warn("event exceeded max retries. moving to DLQ.", "event_id", event.UUID)
					if mErr := o.outboxRepo.MoveToDLQ(txCtx, event, "MAX_RETRIES_EXCEEDED"); mErr != nil {
						o.logger.Error("failed to move max retries event to DLQ", "event_id", event.UUID, "error", mErr)
						return mErr
					}
					continue
				}

				o.logger.Warn("transient error publishing event. incrementing retry", "event_id", event.UUID, "error", pubErr)
				if mErr := o.outboxRepo.IncrementRetry(txCtx, event.UUID); mErr != nil {
					o.logger.Error("failed to increment retry count", "event_id", event.UUID, "error", mErr)
					return mErr
				}
				continue
			}

			if mErr := o.outboxRepo.MarkAsPublished(txCtx, event.UUID); mErr != nil {
				return mErr
			}
		}

		return nil
	})
}
