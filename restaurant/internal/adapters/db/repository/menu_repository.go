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
	"github.com/vterry/food-project/restaurant/internal/core/domain/menu"
)

type SQLMenuRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLMenuRepository(db *sql.DB) *SQLMenuRepository {
	return &SQLMenuRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SQLMenuRepository) Save(ctx context.Context, m *menu.Menu) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	_, err = qtx.GetMenuByID(ctx, m.ID().String())
	if err == sql.ErrNoRows {
		err = qtx.InsertMenu(ctx, sqlc.InsertMenuParams{
			ID:           m.ID().String(),
			RestaurantID: m.RestaurantID().String(),
			Name:         m.Name(),
			IsActive:     sql.NullBool{Bool: m.IsActive(), Valid: true},
		})
	} else if err == nil {
		err = qtx.UpdateMenu(ctx, sqlc.UpdateMenuParams{
			ID:       m.ID().String(),
			Name:     m.Name(),
			IsActive: sql.NullBool{Bool: m.IsActive(), Valid: true},
		})
	}
	if err != nil {
		return fmt.Errorf("failed to save menu: %w", err)
	}

	// Sync items
	if err := qtx.ClearMenuItems(ctx, m.ID().String()); err != nil {
		return err
	}

	for _, item := range m.Items() {
		err = qtx.InsertMenuItem(ctx, sqlc.InsertMenuItemParams{
			ID:            item.ID().String(),
			MenuID:        m.ID().String(),
			Name:          item.Name(),
			Description:   sql.NullString{String: item.Description(), Valid: true},
			PriceAmount:   item.Price().Amount(),
			PriceCurrency: item.Price().Currency(),
			Category:      item.Category(),
			IsAvailable:   sql.NullBool{Bool: item.IsAvailable(), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to save menu item %s: %w", item.Name(), err)
		}
	}

	// Outbox
	correlationID := ctxutil.GetCorrelationID(ctx)
	for _, event := range m.Events() {
		payload, _ := json.Marshal(event)
		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "Menu",
			AggregateID:   m.ID().String(),
			EventType:     event.EventType(),
			Payload:       payload,
			CorrelationID: correlationID,
		})
		if err != nil {
			return err
		}
	}
	m.ClearEvents()

	return tx.Commit()
}

func (r *SQLMenuRepository) FindByID(ctx context.Context, id vo.ID) (*menu.Menu, error) {
	row, err := r.q.GetMenuByID(ctx, id.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapToDomain(ctx, row)
}

func (r *SQLMenuRepository) FindActiveByRestaurantID(ctx context.Context, restaurantID vo.ID) (*menu.Menu, error) {
	row, err := r.q.GetActiveMenuByRestaurantID(ctx, restaurantID.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.mapToDomain(ctx, row)
}

func (r *SQLMenuRepository) mapToDomain(ctx context.Context, row sqlc.Menu) (*menu.Menu, error) {
	m := menu.NewMenu(vo.NewID(row.ID), vo.NewID(row.RestaurantID), row.Name)
	if row.IsActive.Bool {
		m.Activate()
	}
	m.ClearEvents()

	itemRows, err := r.q.ListMenuItemsByMenuID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	for _, ir := range itemRows {
		price, _ := vo.NewMoney(ir.PriceAmount, ir.PriceCurrency)
		item := menu.NewMenuItem(vo.NewID(ir.ID), ir.Name, ir.Description.String, price, ir.Category)
		item.SetAvailability(ir.IsAvailable.Bool)
		m.AddItem(item)
	}

	return m, nil
}
