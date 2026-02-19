package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

func TestOutboxRepository_FindUnpublishedEvents(t *testing.T) {
	repo := NewOutboxRepository(testDB)

	orderedUUIDs := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}

	tests := []struct {
		name       string
		seed       func(t *testing.T)
		limit      int
		wantCount  int
		wantErr    bool
		assertFunc func(t *testing.T, events []output.OutboxEvent)
	}{
		{
			name:      "return empty when there are no event",
			seed:      func(t *testing.T) {},
			limit:     10,
			wantCount: 0,
		},
		{
			name: "return only unpublished events",
			seed: func(t *testing.T) {
				payload := []byte(`{"teste": true}`)
				seedUnpublishedEvent(t, uuid.New().String(), uuid.New().String(), "AggTypeTest", "EventTypeTest", payload)
				seedUnpublishedEvent(t, uuid.New().String(), uuid.New().String(), "AggTypeTest", "EventTypeTest", payload)
				seedPublishedEvent(t, uuid.New().String(), uuid.New().String(), "AggTypeTest", "EventTypeTest", payload)
			},

			limit:     10,
			wantCount: 2,
		},
		{
			name: "respect limit",
			seed: func(t *testing.T) {
				for i := 0; i < 5; i++ {
					payload := []byte(`{"teste": true}`)
					seedUnpublishedEvent(t, uuid.New().String(), uuid.New().String(), "AggTypeTest", "EventTypeTest", payload)
				}
			},

			limit:     3,
			wantCount: 3,
		},
		{
			name: "return events ordered by ASC",
			seed: func(t *testing.T) {
				for i := 0; i < 3; i++ {
					payload := []byte(`{"teste": true}`)
					seedUnpublishedEvent(t, orderedUUIDs[i], uuid.New().String(), "AggTypeTest", "EventTypeTest", payload)
				}
			},

			limit:     10,
			wantCount: 3,
			assertFunc: func(t *testing.T, events []output.OutboxEvent) {
				for i := range events {
					assert.Equal(t, events[i].UUID.String(), orderedUUIDs[i])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(testDB)
			tt.seed(t)

			events, err := repo.FindUnpublishedEvents(context.Background(), tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, events, tt.wantCount)

			if tt.assertFunc != nil {
				tt.assertFunc(t, events)
			}

		})
	}
}

func TestOutboxRepository_MarkAsPublished(t *testing.T) {
	repo := NewOutboxRepository(testDB)

	tests := []struct {
		name       string
		seed       func(t *testing.T) uuid.UUID
		wantErr    bool
		assertFunc func(t *testing.T)
	}{{
		name: "non existing uuid does not return error",
		seed: func(t *testing.T) uuid.UUID {
			return uuid.New()
		},
		wantErr: false,
	},
		{
			name: "mark existing event as published",
			seed: func(t *testing.T) uuid.UUID {
				eventId := uuid.New()
				payload := []byte(`{"teste": true}`)
				seedUnpublishedEvent(t, eventId.String(), uuid.New().String(), "AggTest", "TestEvent", payload)
				return eventId
			},
			wantErr: false,
			assertFunc: func(t *testing.T) {
				event, err := repo.FindUnpublishedEvents(context.Background(), 10)
				assert.NoError(t, err)
				assert.Empty(t, event)
			},
		},
		{
			name: "marking already published event is idempotent",
			seed: func(t *testing.T) uuid.UUID {
				eventId := uuid.New()
				payload := []byte(`{"teste": true}`)
				seedPublishedEvent(t, eventId.String(), uuid.New().String(), "AggTest", "TestEvent", payload)
				return eventId
			},
			wantErr: false,
			assertFunc: func(t *testing.T) {
				event, err := repo.FindUnpublishedEvents(context.Background(), 10)
				assert.NoError(t, err)
				assert.Empty(t, event)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(testDB)
			eventID := tt.seed(t)

			err := repo.MarkAsPublished(context.Background(), eventID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.assertFunc != nil {
				tt.assertFunc(t)
			}
		})
	}
}

func seedUnpublishedEvent(t *testing.T, uuid, aggId, aggType, eventType string, payload []byte) {
	t.Helper()
	_, err := testDB.ExecContext(context.Background(), QueryInsertOutboxEvent, uuid, aggId, aggType, eventType, payload, time.Now())
	require.NoError(t, err)
}

func seedPublishedEvent(t *testing.T, uuid, aggId, aggType, eventType string, payload []byte) {
	t.Helper()
	const q = `INSERT INTO outbox_events (uuid, aggregate_id, aggregate_type, type, payload, occurred_on, published_at) VALUES (?,?,?,?,?,?,?)`
	_, err := testDB.ExecContext(context.Background(), q, uuid, aggId, aggType, eventType, payload, time.Now(), time.Now())
	require.NoError(t, err)
}
