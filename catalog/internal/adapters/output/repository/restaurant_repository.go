package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository/dao"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

var _ output.RestaurantRepository = (*RestaurantRepository)(nil)

type RestaurantRepository struct {
	db *sql.DB
}

func NewRestaurantRepository(db *sql.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) Save(ctx context.Context, agg *restaurant.Restaurant) error {
	executor := getExecutor(ctx, r.db)

	var activeMenuUUID interface{}
	activeMenuID := agg.ActiveMenuID()
	if !activeMenuID.IsZero() {
		activeMenuUUID = activeMenuID.String()
	} else {
		activeMenuUUID = nil
	}

	addr := agg.Address()
	_, err := executor.ExecContext(ctx, QueryUpsertRestaurant,
		agg.ID().String(),
		agg.Name(),
		addr.Street(),
		addr.Number(),
		addr.Complement(),
		addr.Neighborhood(),
		addr.City(),
		addr.State(),
		addr.ZipCode(),
		agg.Status().String(),
		activeMenuUUID,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert restaurant: %w", err)
	}

	for _, event := range agg.PullEvent() {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		_, err = executor.ExecContext(ctx, QueryInsertOutboxEvent,
			event.EventID().String(),
			agg.ID().String(),
			"Restaurant",
			event.EventName(),
			payload,
			event.OccurredOn())

		if err != nil {
			return fmt.Errorf("failed to save outbox event: %w", err)
		}
	}

	return nil
}

func (r *RestaurantRepository) FindById(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
	row := r.db.QueryRowContext(ctx, QueryFindRestaurantByUUID, id.String())

	var dao dao.RestaurantDAO
	err := row.Scan(
		&dao.ID,
		&dao.UUID,
		&dao.Name,
		&dao.AddressStreet,
		&dao.AddressNumber,
		&dao.AddressCompl,
		&dao.AddressNeigh,
		&dao.AddressCity,
		&dao.AddressState,
		&dao.AddressZipCode,
		&dao.Status,
		&dao.ActiveMenuID,
		&dao.CreatedAt,
		&dao.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, output.ErrEntityNotFound
		}
		return nil, err
	}

	return r.toDomain(dao)
}

func (r *RestaurantRepository) FindAll(ctx context.Context) ([]*restaurant.Restaurant, error) {
	rows, err := r.db.QueryContext(ctx, QueryFindAllRestaurants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*restaurant.Restaurant

	for rows.Next() {
		var dao dao.RestaurantDAO
		err := rows.Scan(
			&dao.ID,
			&dao.UUID,
			&dao.Name,
			&dao.AddressStreet,
			&dao.AddressNumber,
			&dao.AddressCompl,
			&dao.AddressNeigh,
			&dao.AddressCity,
			&dao.AddressState,
			&dao.AddressZipCode,
			&dao.Status,
			&dao.ActiveMenuID,
			&dao.CreatedAt,
			&dao.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		agg, err := r.toDomain(dao)
		if err != nil {
			return nil, err
		}
		result = append(result, agg)
	}

	return result, nil
}

func (r *RestaurantRepository) toDomain(dao dao.RestaurantDAO) (*restaurant.Restaurant, error) {
	rID, err := valueobjects.ParseRestaurantId(dao.UUID)
	if err != nil {
		return nil, err
	}

	status, err := enums.ParseRestaurantStatus(dao.Status)
	if err != nil {
		return nil, err
	}

	address, err := valueobjects.NewAddress(
		dao.AddressStreet,
		dao.AddressNumber,
		dao.AddressCompl,
		dao.AddressNeigh,
		dao.AddressCity,
		dao.AddressState,
		dao.AddressZipCode,
	)
	if err != nil {
		return nil, err
	}

	var activeMenuID valueobjects.MenuID
	if dao.ActiveMenuID != nil {
		activeMenuID, err = valueobjects.ParseMenuId(*dao.ActiveMenuID)
		if err != nil {
			return nil, err
		}
	}

	return restaurant.Restore(rID, dao.Name, address, status, activeMenuID), nil
}
