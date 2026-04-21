package base

import (
	"time"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

// DomainEvent is the interface that all domain events must implement.
type DomainEvent interface {
	EventID() vo.ID
	OccurredAt() time.Time
	EventType() string
}

// AggregateRoot is the interface that all aggregate roots must implement.
type AggregateRoot interface {
	ID() vo.ID
	AddEvent(event DomainEvent)
	Events() []DomainEvent
	ClearEvents()
}

// BaseAggregateRoot is a helper struct to implement AggregateRoot interface.
type BaseAggregateRoot struct {
	id     vo.ID
	events []DomainEvent
}

func (b *BaseAggregateRoot) ID() vo.ID {
	return b.id
}

func (b *BaseAggregateRoot) SetID(id vo.ID) {
	b.id = id
}

func (b *BaseAggregateRoot) AddEvent(event DomainEvent) {
	b.events = append(b.events, event)
}

func (b *BaseAggregateRoot) Events() []DomainEvent {
	return b.events
}

func (b *BaseAggregateRoot) ClearEvents() {
	b.events = nil
}

// BaseDomainEvent is a helper struct to implement DomainEvent interface.
type BaseDomainEvent struct {
	id         vo.ID
	occurredAt time.Time
}

func NewBaseDomainEvent(id vo.ID) BaseDomainEvent {
	return BaseDomainEvent{
		id:         id,
		occurredAt: time.Now(),
	}
}

func (b BaseDomainEvent) EventID() vo.ID {
	return b.id
}

func (b BaseDomainEvent) OccurredAt() time.Time {
	return b.occurredAt
}
