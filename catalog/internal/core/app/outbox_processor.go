package app

import (
	"context"
	"log"
	"time"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

type OutboxProcessor struct {
	outboxRepo   output.OutboxRepository
	publisher    output.EventPublisher
	uow          output.UnitOfWork
	pollInterval time.Duration
	batchSize    int
}

func NewOutboxProcessor(repo output.OutboxRepository, pub output.EventPublisher, uow output.UnitOfWork, pollInterval time.Duration, batchSize int) *OutboxProcessor {

	return &OutboxProcessor{
		outboxRepo:   repo,
		publisher:    pub,
		uow:          uow,
		pollInterval: pollInterval,
		batchSize:    batchSize,
	}
}

func (o *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(o.pollInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("OutboxProcessor stopped due to context cancellation")
				return
			case <-ticker.C:
				if err := o.processBatch(ctx); err != nil {
					log.Printf("OutboxProcessor errored processing batch: %v", err)
				}
			}
		}
	}()
}

func (o *OutboxProcessor) processBatch(ctx context.Context) error {
	var eventsToProcess []output.OutboxEvent

	err := o.uow.Run(ctx, func(txCtx context.Context) error {
		events, err := o.outboxRepo.FindUnpublishedEvents(txCtx, o.batchSize)
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
			Payload:       event.Payload,
			OcurredAt:     event.OccurredOn.UnixMilli(),
		}

		if pubErr := o.publisher.Publish(ctx, msg); pubErr != nil {
			return pubErr
		}

		if mErr := o.uow.Run(ctx, func(txCtx context.Context) error {
			return o.outboxRepo.MarkAsPublished(txCtx, event.UUID)
		}); mErr != nil {
			return mErr
		}
	}

	return nil
}
