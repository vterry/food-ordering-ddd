package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

type OutboxProcessor struct {
	outboxRepo   output.OutboxRepository
	publisher    output.EventPublisher
	uow          output.UnitOfWork
	processorID  string
	pollInterval time.Duration
	batchSize    int
	retryCount   int
	logger       *slog.Logger
}

func NewOutboxProcessor(repo output.OutboxRepository, pub output.EventPublisher, uow output.UnitOfWork, pollInterval time.Duration, batchSize int, retryCount int, logger *slog.Logger) *OutboxProcessor {
	hostname, _ := os.Hostname()
	processorID := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	return &OutboxProcessor{
		outboxRepo:   repo,
		publisher:    pub,
		uow:          uow,
		processorID:  processorID,
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
	var eventsToProcess []output.OutboxEvent

	err := o.uow.Run(ctx, func(txCtx context.Context) error {
		events, err := o.outboxRepo.ClaimAndFindEvents(txCtx, o.processorID, o.batchSize, output.DefaultClaimTTL)
		if err != nil {
			return err
		}
		eventsToProcess = events
		return nil
	})

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
				if dbErr := o.uow.Run(ctx, func(txCtx context.Context) error {
					return o.outboxRepo.MoveToDLQ(txCtx, event, pubErr.Error())
				}); dbErr != nil {
					o.logger.Error("failed to move unrecoverable event to DLQ, aborting batch", "event_id", event.UUID, "error", dbErr)
					return dbErr
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
				if dbErr := o.uow.Run(ctx, func(txCtx context.Context) error {
					return o.outboxRepo.MoveToDLQ(txCtx, event, "MAX_RETRIES_EXCEEDED")
				}); dbErr != nil {
					o.logger.Error("failed to move max-retries event to DLQ, aborting batch", "event_id", event.UUID, "error", dbErr)
					return dbErr
				}
				continue
			}

			o.logger.Warn("transient error publishing event. incrementing retry", "event_id", event.UUID, "error", pubErr)
			if dbErr := o.uow.Run(ctx, func(txCtx context.Context) error {
				return o.outboxRepo.IncrementRetry(txCtx, event.UUID)
			}); dbErr != nil {
				o.logger.Error("failed to increment retry count, aborting batch", "event_id", event.UUID, "error", dbErr)
				return dbErr
			}
			continue
		}

		if dbErr := o.uow.Run(ctx, func(txCtx context.Context) error {
			return o.outboxRepo.MarkAsPublished(txCtx, event.UUID)
		}); dbErr != nil {
			o.logger.Error("failed to mark event as published, aborting batch", "event_id", event.UUID, "error", dbErr)
			return dbErr
		}
	}

	return nil
}
