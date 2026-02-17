package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository/dao"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

var _ output.OutboxRepository = (*OutboxRepository)(nil)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (o *OutboxRepository) FindUnpublishedEvents(ctx context.Context, limit int) ([]output.OutboxEvent, error) {
	rows, err := o.db.QueryContext(ctx, QueryFindUnpublishedEvents, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]output.OutboxEvent, 0)
	for rows.Next() {
		var eventDao dao.OutboxEventDAO
		err := rows.Scan(
			&eventDao.ID,
			&eventDao.UUID,
			&eventDao.AggregateID,
			&eventDao.AggregateType,
			&eventDao.EventType,
			&eventDao.Payload,
			&eventDao.OccurredOn,
		)

		if err != nil {
			return nil, err
		}

		eventID, err := uuid.Parse(eventDao.UUID)
		if err != nil {
			return nil, err
		}

		events = append(events, output.OutboxEvent{
			UUID:          eventID,
			AggregateID:   eventDao.AggregateID,
			AggregateType: eventDao.AggregateType,
			EventType:     eventDao.EventType,
			Payload:       eventDao.Payload,
			OccurredOn:    eventDao.OccurredOn,
		})
	}

	return events, nil
}

func (o *OutboxRepository) MarkAsPublished(ctx context.Context, eventID uuid.UUID) error {
	_, err := o.db.ExecContext(ctx, QueryMarkAsPublished, eventID.String())
	return err
}
