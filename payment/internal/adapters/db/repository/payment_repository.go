package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vterry/food-project/payment/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

type paymentRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewPaymentRepository(db *sql.DB) *paymentRepository {
	return &paymentRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *paymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	// Check if payment already exists
	_, err = qtx.GetPaymentByID(ctx, p.ID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create new payment
			err = qtx.CreatePayment(ctx, sqlc.CreatePaymentParams{
				ID:        p.ID(),
				OrderID:   p.OrderID(),
				Amount:    p.Amount(),
				Status:    string(p.Status()),
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			})
			if err != nil {
				return fmt.Errorf("failed to create payment: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get payment: %w", err)
		}
	} else {
		// Update existing payment
		err = qtx.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status:    string(p.Status()),
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			ID:        p.ID(),
		})
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	// Save Outbox messages
	events := p.GetEvents()
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Try to get correlation_id from context, otherwise generate a new one
		correlationID := r.getCorrelationID(ctx)

		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "Payment",
			AggregateID:   p.ID(),
			EventType:     fmt.Sprintf("%T", event),
			Payload:       payload,
			CorrelationID: correlationID,
			CreatedAt:     time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert outbox message: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, id string) (*payment.Payment, error) {
	row, err := r.q.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by id: %w", err)
	}

	return r.mapToDomain(row), nil
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	row, err := r.q.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by order id: %w", err)
	}

	return r.mapToDomain(row), nil
}

func (r *paymentRepository) mapToDomain(row sqlc.Payment) *payment.Payment {
	return payment.MapFromPersistence(
		row.ID,
		row.OrderID,
		row.Amount,
		payment.Status(row.Status),
		row.UpdatedAt.Time,
	)
}

func (r *paymentRepository) getCorrelationID(ctx context.Context) string {
	// Placeholder for correlation ID logic. 
	// In a real system, this would come from a context util.
	return uuid.New().String()
}
