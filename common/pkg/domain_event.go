package common

import (
	"time"

	"github.com/google/uuid"
)

type EventMetadata struct {
	eventID       uuid.UUID
	occurredOn    time.Time
	correlationID uuid.UUID
}

type DomainEvent interface {
	EventID() uuid.UUID
	EventName() string
	AggregateID() string
	OccurredOn() time.Time
	CorrelationID() uuid.UUID
}

type BaseEvent struct {
	EventMetadata
	aggregateID   string
	aggregateType string
	eventName     string
}

func NewBaseEvent(eventName, aggregateID string) BaseEvent {
	return BaseEvent{
		EventMetadata: newEventMetadata(),
		eventName:     eventName,
		aggregateID:   aggregateID,
	}
}

func (e BaseEvent) AggregateID() string {
	return e.aggregateID
}

func (e BaseEvent) EventName() string {
	return e.eventName
}

func (e BaseEvent) EventID() uuid.UUID {
	return e.EventMetadata.eventID
}

func (e BaseEvent) OccurredOn() time.Time {
	return e.EventMetadata.occurredOn
}

func (e BaseEvent) CorrelationID() uuid.UUID {
	return e.EventMetadata.correlationID
}

func (e BaseEvent) WithCorrelationID(correlationID uuid.UUID) BaseEvent {
	e.correlationID = correlationID
	return e
}

func newEventMetadata() EventMetadata {
	return EventMetadata{
		eventID:    uuid.New(),
		occurredOn: time.Now(),
	}
}
