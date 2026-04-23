package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/vterry/food-project/ordering/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/ordering/internal/core/domain/saga"
)

type SQLSagaRepository struct {
	queries *sqlc.Queries
}

func NewSQLSagaRepository(db *sql.DB) *SQLSagaRepository {
	return &SQLSagaRepository{
		queries: sqlc.New(db),
	}
}

func (r *SQLSagaRepository) Save(ctx context.Context, s *saga.SagaState) error {
	dataJSON, err := json.Marshal(s.Data())
	if err != nil {
		return err
	}

	return r.queries.SaveSagaState(ctx, sqlc.SaveSagaStateParams{
		OrderID:     s.OrderID(),
		CurrentStep: s.CurrentStep(),
		Status:      string(s.Status()),
		DataJson:    dataJSON,
		UpdatedAt:   s.UpdatedAt(),
	})
}

func (r *SQLSagaRepository) FindByOrderID(ctx context.Context, orderID string) (*saga.SagaState, error) {
	dbSaga, err := r.queries.GetSagaState(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return saga.MapFromPersistence(
		dbSaga.OrderID,
		dbSaga.CurrentStep,
		saga.SagaStatus(dbSaga.Status),
		dbSaga.DataJson,
		dbSaga.UpdatedAt,
	)
}
