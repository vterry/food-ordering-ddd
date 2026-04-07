package dao

import (
	"time"
)

type OutboxEventDAO struct {
	ID            int64      `db:"id"`
	UUID          string     `db:"uuid"`
	AggregateID   string     `db:"aggregate_id"`
	AggregateType string     `db:"aggregate_type"`
	EventType     string     `db:"type"`
	Payload       []byte     `db:"payload"`
	OccurredOn    time.Time  `db:"occurred_on"`
	RetryCount    int        `db:"retry_count"`
	ClaimedBy     *string    `db:"claimed_by"`
	ClaimedAt     *time.Time `db:"claimed_at"`
}
