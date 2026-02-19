package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository/dao"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

var _ output.MenuRepository = (*MenuRepository)(nil)

type MenuRepository struct {
	db     *sql.DB
	outbox output.OutboxRepository
}

type menuBasket struct {
	Name         string
	RestaurantID string
	Status       enums.MenuStatus
	Categories   map[string]*categoryBasket
}

type categoryBasket struct {
	Name  string
	Items []menu.ItemMenu
}

func NewMenuRepository(db *sql.DB, outbotx output.OutboxRepository) *MenuRepository {
	return &MenuRepository{
		db:     db,
		outbox: outbotx,
	}
}

func (m *MenuRepository) Save(ctx context.Context, menuAgg *menu.Menu) error {
	executor := getExecutor(ctx, m.db)

	_, err := executor.ExecContext(ctx, QueryUpsertMenu,
		menuAgg.String(),
		menuAgg.RestaurantID().String(),
		menuAgg.Name(),
		menuAgg.Status().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert menu: %w", err)
	}

	var menuDbId int64
	err = executor.QueryRowContext(ctx, QueryGetMenuIDByUUID, menuAgg.String()).Scan(&menuDbId)
	if err != nil {
		return fmt.Errorf("failed to get menu db id: %w", err)
	}

	if _, err := executor.ExecContext(ctx, QueryDeleteCategoriesByMenuID, menuDbId); err != nil {
		return fmt.Errorf("failed to cleanup categories: %w", err)
	}

	for _, cat := range menuAgg.Categories() {
		res, err := executor.ExecContext(ctx, QueryInsertCategory,
			cat.CategoryID.String(),
			menuDbId,
			cat.Name(),
			"",
		)

		if err != nil {
			return fmt.Errorf("failed to insert category: %w", err)
		}

		catDbID, _ := res.LastInsertId()

		for _, item := range cat.Items() {
			_, err := executor.ExecContext(ctx, QueryInsertItem,
				item.ItemID.String(),
				catDbID,
				item.Name(),
				item.Description(),
				item.BasePrice().Amount(),
				item.Status().String())

			if err != nil {
				return fmt.Errorf("failed to insert item: %w", err)
			}
		}
	}

	if err := m.outbox.SaveEvents(ctx, menuAgg.String(), "Menu", menuAgg.PullEvent()); err != nil {
		return err
	}

	return nil
}

func (m *MenuRepository) FindById(ctx context.Context, menuId valueobjects.MenuID) (*menu.Menu, error) {
	rows, err := m.db.QueryContext(ctx, QueryFindFullMenuByUUID, menuId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus, err := scanMenuRows(rows)
	if err != nil {
		return nil, err
	}

	if len(menus) == 0 {
		return nil, output.ErrEntityNotFound
	}

	return menus[0], nil
}

func (m *MenuRepository) FindByRestaurantId(ctx context.Context, restaurantId valueobjects.RestaurantID) ([]*menu.Menu, error) {
	rows, err := m.db.QueryContext(ctx, QueryFindAllMenusByRestaurantID, restaurantId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuRows(rows)
}

func (m *MenuRepository) FindActiveMenuByRestaurantId(ctx context.Context, restaurantId valueobjects.RestaurantID) (*menu.Menu, error) {
	rows, err := m.db.QueryContext(ctx, QueryFindActiveMenusByRestaurantID, restaurantId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus, err := scanMenuRows(rows)
	if err != nil {
		return nil, err
	}

	if len(menus) == 0 {
		return nil, nil
	}

	return menus[0], nil
}

func scanMenuRows(rows *sql.Rows) ([]*menu.Menu, error) {
	baskets := make(map[string]*menuBasket)
	var order []string

	for rows.Next() {
		var row dao.MenuCompositeDAO

		err := rows.Scan(
			&row.MenuID, &row.MenuUUID, &row.RestaurantID, &row.MenuName, &row.MenuStatus,
			&row.CategoryID, &row.CategoryUUID, &row.CategoryName,
			&row.ItemID, &row.ItemUUID, &row.ItemName, &row.ItemDesc, &row.ItemPrice, &row.ItemStatus,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := baskets[row.MenuUUID]; !exists {
			menuStatus, err := enums.ParseMenuStatus(row.MenuStatus)
			if err != nil {
				return nil, fmt.Errorf("corrupted menu status %q for menu %s: %w", row.MenuStatus, row.MenuUUID, err)
			}

			baskets[row.MenuUUID] = &menuBasket{
				Name:         row.MenuName,
				RestaurantID: row.RestaurantID,
				Status:       menuStatus,
				Categories:   make(map[string]*categoryBasket),
			}
			order = append(order, row.MenuUUID)
		}

		if err := processChildren(row, baskets[row.MenuUUID].Categories); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*menu.Menu, 0, len(baskets))

	for _, uuid := range order {
		b := baskets[uuid]

		mID, err := valueobjects.ParseMenuId(uuid)
		if err != nil {
			return nil, fmt.Errorf("corrupted menu id %q: %w", uuid, err)
		}

		rID, err := valueobjects.ParseRestaurantId(b.RestaurantID)
		if err != nil {
			return nil, fmt.Errorf("corrupted restaurant id %q: %w", b.RestaurantID, err)
		}

		categories := make([]menu.Category, 0, len(b.Categories))
		for catUUID, cb := range b.Categories {
			cid, err := valueobjects.ParseCategoryId(catUUID)
			if err != nil {
				return nil, fmt.Errorf("corrupted category id %q: %w", catUUID, err)
			}
			cat := menu.RestoreCategory(cid, cb.Name, cb.Items)
			categories = append(categories, *cat)
		}

		menuObj := menu.Restore(mID, b.Name, rID, b.Status, categories)
		result = append(result, menuObj)
	}

	return result, nil
}

func processChildren(row dao.MenuCompositeDAO, cats map[string]*categoryBasket) error {
	if row.CategoryUUID == nil {
		return nil
	}

	catUUID := *row.CategoryUUID

	if _, exists := cats[catUUID]; !exists {
		cats[catUUID] = &categoryBasket{Name: *row.CategoryName}
	}

	if row.ItemUUID == nil {
		return nil
	}

	iid, err := valueobjects.ParseItemId(*row.ItemUUID)
	if err != nil {
		return fmt.Errorf("corrupted item id %q: %w", *row.ItemUUID, err)
	}

	status, err := enums.ParseItemStatus(*row.ItemStatus)
	if err != nil {
		return fmt.Errorf("corrupted item status %q for item %s: %w", *row.ItemStatus, *row.ItemUUID, err)
	}

	item := menu.RestoreItemMenu(iid, *row.ItemName, *row.ItemDesc, common.NewMoneyFromCents(*row.ItemPrice), status)
	cats[catUUID].Items = append(cats[catUUID].Items, *item)

	return nil
}
