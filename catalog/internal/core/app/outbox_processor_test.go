package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ---- MOCKS ----

type MockOutboxRepository struct {
	SaveEventsFunc            func(ctx context.Context, aggregateID, aggregateType string, events []common.DomainEvent) error
	FindUnpublishedEventsFunc func(ctx context.Context, limit int) ([]output.OutboxEvent, error)
	MarkAsPublishedFunc       func(ctx context.Context, eventID uuid.UUID) error
	IncrementRetryFunc        func(ctx context.Context, eventID uuid.UUID) error
	MoveToDLQFunc             func(ctx context.Context, event output.OutboxEvent, reason string) error
}

func (m *MockOutboxRepository) SaveEvents(ctx context.Context, aggregateID, aggregateType string, events []common.DomainEvent) error {
	if m.SaveEventsFunc != nil {
		return m.SaveEventsFunc(ctx, aggregateID, aggregateType, events)
	}
	return nil
}

func (m *MockOutboxRepository) FindUnpublishedEvents(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
	if m.FindUnpublishedEventsFunc != nil {
		return m.FindUnpublishedEventsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockOutboxRepository) MarkAsPublished(ctx context.Context, eventID uuid.UUID) error {
	if m.MarkAsPublishedFunc != nil {
		return m.MarkAsPublishedFunc(ctx, eventID)
	}
	return nil
}

func (m *MockOutboxRepository) IncrementRetry(ctx context.Context, eventID uuid.UUID) error {
	if m.IncrementRetryFunc != nil {
		return m.IncrementRetryFunc(ctx, eventID)
	}
	return nil
}

func (m *MockOutboxRepository) MoveToDLQ(ctx context.Context, event output.OutboxEvent, reason string) error {
	if m.MoveToDLQFunc != nil {
		return m.MoveToDLQFunc(ctx, event, reason)
	}
	return nil
}

type MockEventPublisher struct {
	PublishFunc func(ctx context.Context, message output.EventMessage) error
}

func (m *MockEventPublisher) Publish(ctx context.Context, message output.EventMessage) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, message)
	}
	return nil
}

type MockUnitOfWork struct{}

func (m *MockUnitOfWork) Run(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// ---- TESTS ----

func TestOutboxProcessor_ProcessBatch(t *testing.T) {
	mockUoW := &MockUnitOfWork{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Caminho Feliz: Publica e marca como publicado", func(t *testing.T) {
		eventID := uuid.New()
		events := []output.OutboxEvent{
			{UUID: eventID, RetryCount: 0},
		}

		mockRepo := &MockOutboxRepository{
			FindUnpublishedEventsFunc: func(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
				return events, nil
			},
			MarkAsPublishedFunc: func(ctx context.Context, id uuid.UUID) error {
				assert.Equal(t, eventID, id)
				return nil
			},
		}

		mockPublisher := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, message output.EventMessage) error {
				assert.Equal(t, eventID.String(), message.EventID)
				return nil
			},
		}

		processor := NewOutboxProcessor(mockRepo, mockPublisher, mockUoW, 1*time.Second, 10, 3, logger)
		err := processor.processBatch(context.Background())

		assert.NoError(t, err)
	})

	t.Run("Erro Irrecuperavel: Move para DLQ e continua batch", func(t *testing.T) {
		eventID := uuid.New()
		events := []output.OutboxEvent{
			{UUID: eventID, RetryCount: 0},
		}

		var movedToDLQ bool
		mockRepo := &MockOutboxRepository{
			FindUnpublishedEventsFunc: func(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
				return events, nil
			},
			MoveToDLQFunc: func(ctx context.Context, event output.OutboxEvent, reason string) error {
				movedToDLQ = true
				assert.Equal(t, eventID, event.UUID)
				assert.Contains(t, reason, "poison")
				return nil
			},
		}

		mockPublisher := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, message output.EventMessage) error {
				return common.NewUnrecoverableErr(errors.New("poison message format"))
			},
		}

		processor := NewOutboxProcessor(mockRepo, mockPublisher, mockUoW, 1*time.Second, 10, 3, logger)
		err := processor.processBatch(context.Background())

		assert.NoError(t, err) // Batch continua normal!
		assert.True(t, movedToDLQ)
	})

	t.Run("Erro de Infra: Interrompe o batch", func(t *testing.T) {
		eventID := uuid.New()
		events := []output.OutboxEvent{
			{UUID: eventID, RetryCount: 0},
		}

		mockRepo := &MockOutboxRepository{
			FindUnpublishedEventsFunc: func(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
				return events, nil
			},
		}

		mockPublisher := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, message output.EventMessage) error {
				return common.NewInfraConnectionErr(errors.New("rabbitmq connection refused"))
			},
		}

		processor := NewOutboxProcessor(mockRepo, mockPublisher, mockUoW, 1*time.Second, 10, 3, logger)
		err := processor.processBatch(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rabbitmq connection refused")
	})

	t.Run("Erro Transitorio: Incrementa Retry", func(t *testing.T) {
		eventID := uuid.New()
		events := []output.OutboxEvent{
			{UUID: eventID, RetryCount: 1}, // Ainda não estourou o limite de 3
		}

		var retryIncremented bool
		mockRepo := &MockOutboxRepository{
			FindUnpublishedEventsFunc: func(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
				return events, nil
			},
			IncrementRetryFunc: func(ctx context.Context, id uuid.UUID) error {
				retryIncremented = true
				assert.Equal(t, eventID, id)
				return nil
			},
		}

		mockPublisher := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, message output.EventMessage) error {
				return errors.New("timeout temporary") // Erro comum, bate no else if e incrementa
			},
		}

		processor := NewOutboxProcessor(mockRepo, mockPublisher, mockUoW, 1*time.Second, 10, 3, logger)
		err := processor.processBatch(context.Background())

		assert.NoError(t, err) // Continua o loop do batch (não aborta)
		assert.True(t, retryIncremented)
	})

	t.Run("Erro Transitorio: Maximo de Retries e Move Pra DLQ", func(t *testing.T) {
		eventID := uuid.New()
		events := []output.OutboxEvent{
			{UUID: eventID, RetryCount: 3}, // Já atingiu limite igual max
		}

		var movedToDLQ bool
		mockRepo := &MockOutboxRepository{
			FindUnpublishedEventsFunc: func(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
				return events, nil
			},
			MoveToDLQFunc: func(ctx context.Context, event output.OutboxEvent, reason string) error {
				movedToDLQ = true
				assert.Equal(t, eventID, event.UUID)
				assert.Equal(t, "MAX_RETRIES_EXCEEDED", reason)
				return nil
			},
		}

		mockPublisher := &MockEventPublisher{
			PublishFunc: func(ctx context.Context, message output.EventMessage) error {
				return errors.New("timeout temporary again")
			},
		}

		processor := NewOutboxProcessor(mockRepo, mockPublisher, mockUoW, 1*time.Second, 10, 3, logger)
		err := processor.processBatch(context.Background())

		assert.NoError(t, err)
		assert.True(t, movedToDLQ)
	})
}
