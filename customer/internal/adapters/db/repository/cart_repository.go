package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
)

type SQLCartRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLCartRepository(db *sql.DB) *SQLCartRepository {
	return &SQLCartRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SQLCartRepository) Save(ctx context.Context, c *cart.Cart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	// 1. Upsert Cart
	err = qtx.InsertCart(ctx, sqlc.InsertCartParams{
		CustomerID:   c.CustomerID().String(),
		RestaurantID: sql.NullString{String: c.RestaurantID().String(), Valid: !c.RestaurantID().IsEmpty()},
	})
	if err != nil {
		return fmt.Errorf("failed to save cart: %w", err)
	}

	// 2. Sync Items
	if err := qtx.DeleteCartItems(ctx, c.CustomerID().String()); err != nil {
		return fmt.Errorf("failed to clear cart items: %w", err)
	}

	for _, item := range c.Items() {
		err = qtx.InsertCartItem(ctx, sqlc.InsertCartItemParams{
			CustomerID:    c.CustomerID().String(),
			ProductID:     item.ProductID().String(),
			Name:          item.Name(),
			PriceAmount:   item.Price().Amount(),
			PriceCurrency: item.Price().Currency(),
			Quantity:      int32(item.Quantity()),
			Observation:   sql.NullString{String: item.Observation(), Valid: item.Observation() != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to save cart item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *SQLCartRepository) FindByCustomerID(ctx context.Context, customerID vo.ID) (*cart.Cart, error) {
	cartRow, err := r.q.GetCartByCustomerID(ctx, customerID.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	c := cart.NewCart(vo.NewID("cart-"+customerID.String()), customerID)
	
	itemRows, err := r.q.ListCartItemsByCustomerID(ctx, customerID.String())
	if err != nil {
		return nil, err
	}

	restID := vo.NewID(cartRow.RestaurantID.String)
	for _, ir := range itemRows {
		price, _ := vo.NewMoney(ir.PriceAmount, ir.PriceCurrency)
		item := cart.NewCartItem(
			vo.NewID(ir.ProductID),
			ir.Name,
			price,
			int(ir.Quantity),
			ir.Observation.String,
		)
		_ = c.AddItem(restID, item)
	}
	c.ClearEvents()

	return c, nil
}

func (r *SQLCartRepository) Delete(ctx context.Context, customerID vo.ID) error {
	return r.q.DeleteCart(ctx, customerID.String())
}
