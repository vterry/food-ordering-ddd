package repository

import (
	"context"
	"database/sql"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

var _ output.UnitOfWork = (*UnitOfWork)(nil)

type txKey struct {
}

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

func (u *UnitOfWork) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		return err
	}

	return tx.Commit()
}

func getExecutor(ctx context.Context, defaultDB *sql.DB) common.DBExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultDB
}
