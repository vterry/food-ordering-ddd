package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/common/pkg/outbox"
	"github.com/vterry/food-project/delivery/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
)

type MySQLDeliveryRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewMySQLDeliveryRepository(db *sql.DB) *MySQLDeliveryRepository {
	return &MySQLDeliveryRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *MySQLDeliveryRepository) Save(ctx context.Context, d *delivery.Delivery) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// 1. Save Delivery
	courierID := sql.NullString{Valid: false}
	courierName := sql.NullString{Valid: false}
	if d.Courier() != nil {
		courierID = sql.NullString{String: d.Courier().ID, Valid: true}
		courierName = sql.NullString{String: d.Courier().Name, Valid: true}
	}

	err = qtx.SaveDelivery(ctx, sqlc.SaveDeliveryParams{
		ID:           d.ID().String(),
		OrderID:      d.OrderID().String(),
		RestaurantID: d.RestaurantID().String(),
		CustomerID:   d.CustomerID().String(),
		Status:       d.Status().String(),
		Street:       d.Address().Street,
		City:         d.Address().City,
		ZipCode:      d.Address().ZipCode,
		CourierID:    courierID,
		CourierName:  courierName,
	})
	if err != nil {
		return err
	}

	// 2. Save Outbox Messages
	for _, e := range d.Events() {
		payload, _ := json.Marshal(e)
		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "DELIVERY",
			AggregateID:   d.ID().String(),
			EventType:     e.EventType(),
			Payload:       payload,
			CorrelationID: d.CorrelationID().String(),
		})
		if err != nil {
			return err
		}
	}

	d.ClearEvents()

	return tx.Commit()
}

func (r *MySQLDeliveryRepository) FindByID(ctx context.Context, id vo.ID) (*delivery.Delivery, error) {
	dbDel, err := r.queries.GetDeliveryByID(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var courier *delivery.CourierInfo
	if dbDel.CourierID.Valid && dbDel.CourierName.Valid {
		c, _ := delivery.NewCourierInfo(dbDel.CourierID.String, dbDel.CourierName.String)
		courier = &c
	}

	d := delivery.MapFromPersistence(
		vo.NewID(dbDel.ID),
		vo.NewID(dbDel.OrderID),
		vo.NewID(dbDel.RestaurantID),
		vo.NewID(dbDel.CustomerID),
		delivery.NewAddress(dbDel.Street, dbDel.City, dbDel.ZipCode),
		delivery.Status(dbDel.Status),
		courier,
		vo.NewID(""),
		dbDel.CreatedAt,
		dbDel.UpdatedAt,
	)

	return d, nil
}

func (r *MySQLDeliveryRepository) FetchUnpublished(ctx context.Context, limit int) ([]outbox.OutboxMessage, error) {
	msgs, err := r.queries.GetUnpublishedOutboxMessages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]outbox.OutboxMessage, 0, len(msgs))
	for i, m := range msgs {
		if i >= limit {
			break
		}
		result = append(result, outbox.OutboxMessage{
			ID:            m.ID,
			AggregateType: m.AggregateType,
			AggregateID:   m.AggregateID,
			EventType:     m.EventType,
			Payload:       m.Payload,
			CorrelationID: m.CorrelationID,
		})
	}
	return result, nil
}

func (r *MySQLDeliveryRepository) MarkAsPublished(ctx context.Context, id string) error {
	return r.queries.MarkOutboxMessageAsPublished(ctx, id)
}

func (r *MySQLDeliveryRepository) IsMessageProcessed(ctx context.Context, messageID string) (bool, error) {
	return r.queries.IsMessageProcessed(ctx, messageID)
}

func (r *MySQLDeliveryRepository) MarkMessageAsProcessed(ctx context.Context, messageID string) error {
	return r.queries.MarkMessageAsProcessed(ctx, messageID)
}
