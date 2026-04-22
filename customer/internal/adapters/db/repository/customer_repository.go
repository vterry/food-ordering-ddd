package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/adapters/db/sqlc"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/google/uuid"
)

type SQLCustomerRepository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewSQLCustomerRepository(db *sql.DB) *SQLCustomerRepository {
	return &SQLCustomerRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *SQLCustomerRepository) Save(ctx context.Context, c *customer.Customer) error {
	// Usar transação para salvar cliente e endereços atomiticamente
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	// 1. Upsert Customer (Simulado com Get + Insert ou Update)
	_, err = qtx.GetCustomerByID(ctx, c.ID().String())
	if err == sql.ErrNoRows {
		err = qtx.InsertCustomer(ctx, sqlc.InsertCustomerParams{
			ID:    c.ID().String(),
			Name:  c.Name().String(),
			Email: c.Email().String(),
			Phone: c.Phone().String(),
		})
	} else if err == nil {
		err = qtx.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
			ID:    c.ID().String(),
			Name:  c.Name().String(),
			Email: c.Email().String(),
			Phone: c.Phone().String(),
		})
	}
	if err != nil {
		return fmt.Errorf("failed to save customer: %w", err)
	}

	// 2. Sync Addresses (Clear and Insert)
	if err := qtx.ClearAddresses(ctx, c.ID().String()); err != nil {
		return fmt.Errorf("failed to clear old addresses: %w", err)
	}

	for _, addr := range c.Addresses() {
		err = qtx.InsertAddress(ctx, sqlc.InsertAddressParams{
			ID:         addr.ID().String(),
			CustomerID: c.ID().String(),
			Street:     addr.Street(),
			City:       addr.City(),
			ZipCode:    addr.ZipCode(),
			IsDefault:  sql.NullBool{Bool: addr.IsDefault(), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to save address: %w", err)
		}
	}

	// 3. Persist Events to Outbox
	correlationID := ctxutil.GetCorrelationID(ctx)
	for _, event := range c.Events() {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.EventType(), err)
		}

		err = qtx.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
			ID:            uuid.New().String(),
			AggregateType: "Customer",
			AggregateID:   c.ID().String(),
			EventType:     event.EventType(),
			Payload:       payload,
			CorrelationID: correlationID,
		})
		if err != nil {
			return fmt.Errorf("failed to insert outbox message: %w", err)
		}
	}

	c.ClearEvents()

	return tx.Commit()
}

func (r *SQLCustomerRepository) FindByID(ctx context.Context, id vo.ID) (*customer.Customer, error) {
	row, err := r.q.GetCustomerByID(ctx, id.String())
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.mapToDomain(ctx, row)
}

func (r *SQLCustomerRepository) FindByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	row, err := r.q.GetCustomerByEmail(ctx, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.mapToDomain(ctx, row)
}

func (r *SQLCustomerRepository) mapToDomain(ctx context.Context, row sqlc.Customer) (*customer.Customer, error) {
	name, _ := customer.NewName(row.Name)
	email, _ := customer.NewEmail(row.Email)
	phone, _ := customer.NewPhone(row.Phone)
	
	c := customer.NewCustomer(vo.NewID(row.ID), name, email, phone)
	c.ClearEvents() // Mapeamento do banco não deve disparar eventos de "novo registro"

	// Carregar endereços
	addrRows, err := r.q.ListAddressesByCustomerID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	for _, ar := range addrRows {
		addr := customer.NewAddress(
			vo.NewID(ar.ID),
			vo.NewID(ar.CustomerID),
			ar.Street,
			ar.City,
			ar.ZipCode,
			ar.IsDefault.Bool,
		)
		c.AddAddress(addr)
	}

	return c, nil
}
