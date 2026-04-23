package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/ordering/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/ordering/internal/core/domain/order"
)

type SQLOrderRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSQLOrderRepository(db *sql.DB) *SQLOrderRepository {
	return &SQLOrderRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SQLOrderRepository) Save(ctx context.Context, o *order.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// 1. Create Order
	err = qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
		ID:              o.ID().String(),
		CustomerID:      o.CustomerID().String(),
		RestaurantID:    o.RestaurantID().String(),
		Status:          o.Status().String(),
		TotalAmount:     int64(o.TotalAmount().Amount() * 100),
		Currency:        o.TotalAmount().Currency(),
		DeliveryAddress: o.DeliveryAddress(),
		CorrelationID:   o.CorrelationID().String(),
		Version:         int32(o.Version()),
		CreatedAt:       o.CreatedAt(),
		UpdatedAt:       o.UpdatedAt(),
	})
	if err != nil {
		return err
	}

	// 2. Create Items
	for _, item := range o.Items() {
		err = qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			ID:         uuid.New().String(),
			OrderID:    o.ID().String(),
			MenuItemID: item.MenuItemID().String(),
			Name:       item.Name(),
			Quantity:   int32(item.Quantity()),
			UnitPrice:  int64(item.UnitPrice().Amount() * 100),
			Notes:      sql.NullString{String: item.Notes(), Valid: item.Notes() != ""},
		})
		if err != nil {
			return err
		}
	}

	// 3. Save Outbox Messages
	for _, e := range o.Events() {
		payload, _ := json.Marshal(e)
		err = qtx.SaveOutboxMessage(ctx, sqlc.SaveOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "ORDER",
			AggregateID:   o.ID().String(),
			EventType:     e.EventType(),
			Payload:       payload,
			CorrelationID: o.CorrelationID().String(),
			CreatedAt:     time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLOrderRepository) FindByID(ctx context.Context, id string) (*order.Order, error) {
	dbOrder, err := r.queries.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	dbItems, err := r.queries.ListOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}

	items := make([]order.OrderItem, 0, len(dbItems))
	for _, it := range dbItems {
		items = append(items, order.NewOrderItem(
			vo.NewID(it.MenuItemID),
			it.Name,
			int(it.Quantity),
			vo.NewMoneyFromFloat(float64(it.UnitPrice)/100.0, dbOrder.Currency),
			it.Notes.String,
		))
	}

	o := order.MapFromPersistence(
		vo.NewID(dbOrder.ID),
		vo.NewID(dbOrder.CustomerID),
		vo.NewID(dbOrder.RestaurantID),
		items,
		order.OrderStatus(dbOrder.Status),
		vo.NewMoneyFromFloat(float64(dbOrder.TotalAmount)/100.0, dbOrder.Currency),
		dbOrder.DeliveryAddress,
		vo.NewID(dbOrder.CorrelationID),
		int(dbOrder.Version),
		dbOrder.CreatedAt,
		dbOrder.UpdatedAt,
	)

	return o, nil
}

func (r *SQLOrderRepository) UpdateWithVersion(ctx context.Context, o *order.Order, expectedVersion int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// 1. Update Order Status and Version
	err = qtx.UpdateOrderWithVersion(ctx, sqlc.UpdateOrderWithVersionParams{
		Status:    o.Status().String(),
		UpdatedAt: o.UpdatedAt(),
		ID:        o.ID().String(),
		Version:   int32(expectedVersion),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("concurrency error: order %s version mismatch", o.ID())
		}
		return err
	}

	// 2. Save Outbox Messages (Domain Events)
	for _, e := range o.Events() {
		payload, _ := json.Marshal(e)
		err = qtx.SaveOutboxMessage(ctx, sqlc.SaveOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "ORDER",
			AggregateID:   o.ID().String(),
			EventType:     e.EventType(),
			Payload:       payload,
			CorrelationID: o.CorrelationID().String(),
			CreatedAt:     time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
