package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
)

type SQLTicketRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLTicketRepository(db *sql.DB) *SQLTicketRepository {
	return &SQLTicketRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SQLTicketRepository) Save(ctx context.Context, t *ticket.Ticket) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	_, err = qtx.GetTicketByID(ctx, t.ID().String())
	if err == sql.ErrNoRows {
		err = qtx.InsertTicket(ctx, sqlc.InsertTicketParams{
			ID:           t.ID().String(),
			OrderID:      t.OrderID().String(),
			RestaurantID: t.RestaurantID().String(),
			Status:       string(t.Status()),
			RejectReason: sql.NullString{String: "", Valid: false}, // Reason not part of aggregate fields yet
		})
		if err == nil {
			for _, item := range t.Items() {
				err = qtx.InsertTicketItem(ctx, sqlc.InsertTicketItemParams{
					TicketID:  t.ID().String(),
					ProductID: item.ProductID.String(),
					Name:      item.Name,
					Quantity:  int32(item.Quantity),
				})
				if err != nil {
					return err
				}
			}
		}
	} else if err == nil {
		err = qtx.UpdateTicketStatus(ctx, sqlc.UpdateTicketStatusParams{
			ID:           t.ID().String(),
			Status:       string(t.Status()),
			RejectReason: sql.NullString{String: "", Valid: false},
		})
	}
	if err != nil {
		return fmt.Errorf("failed to save ticket: %w", err)
	}

	// Outbox
	correlationID := ctxutil.GetCorrelationID(ctx)
	for _, event := range t.Events() {
		payload, _ := json.Marshal(event)
		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "Ticket",
			AggregateID:   t.ID().String(),
			EventType:     event.EventType(),
			Payload:       payload,
			CorrelationID: correlationID,
		})
		if err != nil {
			return err
		}
	}
	t.ClearEvents()

	return tx.Commit()
}

func (r *SQLTicketRepository) FindByID(ctx context.Context, id vo.ID) (*ticket.Ticket, error) {
	row, err := r.q.GetTicketByID(ctx, id.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	itemRows, err := r.q.ListTicketItemsByTicketID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	items := make([]ticket.TicketItem, 0, len(itemRows))
	for _, ir := range itemRows {
		items = append(items, ticket.TicketItem{
			ProductID: vo.NewID(ir.ProductID),
			Name:      ir.Name,
			Quantity:  int(ir.Quantity),
		})
	}

	return ticket.MapFromPersistence(
		vo.NewID(row.ID),
		vo.NewID(row.OrderID),
		vo.NewID(row.RestaurantID),
		ticket.TicketStatus(row.Status),
		row.RejectReason.String,
		items,
	), nil
}
