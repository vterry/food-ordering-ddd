package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
)

type SQLRestaurantRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLRestaurantRepository(db *sql.DB) *SQLRestaurantRepository {
	return &SQLRestaurantRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SQLRestaurantRepository) Save(ctx context.Context, rest *restaurant.Restaurant) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	hoursJson, err := json.Marshal(rest.OperatingHours())
	if err != nil {
		return fmt.Errorf("failed to marshal operating hours: %w", err)
	}

	_, err = qtx.GetRestaurantByID(ctx, rest.ID().String())
	if err == sql.ErrNoRows {
		err = qtx.InsertRestaurant(ctx, sqlc.InsertRestaurantParams{
			ID:             rest.ID().String(),
			Name:           rest.Name(),
			Street:         rest.Address().Street,
			City:           rest.Address().City,
			ZipCode:        rest.Address().ZipCode,
			OperatingHours: json.RawMessage(hoursJson),
		})
	} else if err == nil {
		err = qtx.UpdateRestaurant(ctx, sqlc.UpdateRestaurantParams{
			ID:             rest.ID().String(),
			Name:           rest.Name(),
			Street:         rest.Address().Street,
			City:           rest.Address().City,
			ZipCode:        rest.Address().ZipCode,
			OperatingHours: json.RawMessage(hoursJson),
		})
	}

	if err != nil {
		return fmt.Errorf("failed to save restaurant: %w", err)
	}

	// Persist Events to Outbox
	if err := r.persistEvents(ctx, qtx, rest.ID().String(), "Restaurant", rest.Events()); err != nil {
		return err
	}
	rest.ClearEvents()

	return tx.Commit()
}

func (r *SQLRestaurantRepository) FindByID(ctx context.Context, id vo.ID) (*restaurant.Restaurant, error) {
	row, err := r.q.GetRestaurantByID(ctx, id.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var hours []restaurant.OperatingPeriod
	if err := json.Unmarshal(row.OperatingHours, &hours); err != nil {
		return nil, fmt.Errorf("failed to unmarshal operating hours: %w", err)
	}

	addr := restaurant.Address{
		Street:  row.Street,
		City:    row.City,
		ZipCode: row.ZipCode,
	}

	rest := restaurant.NewRestaurant(vo.NewID(row.ID), row.Name, addr, hours)
	rest.ClearEvents()
	return rest, nil
}

// Helper to avoid duplication
func (r *SQLRestaurantRepository) persistEvents(ctx context.Context, qtx *sqlc.Queries, aggregateID, aggregateType string, events []base.DomainEvent) error {
	correlationID := ctxutil.GetCorrelationID(ctx)
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return err
		}

		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: aggregateType,
			AggregateID:   aggregateID,
			EventType:     event.EventType(),
			Payload:       payload,
			CorrelationID: correlationID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Note: persistEvents uses interface{} because we have different event types. 
// However, base.DomainEvent is preferred.
