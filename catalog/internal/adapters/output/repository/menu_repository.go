package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	db *sql.DB
}

type menuBasket struct {
	Name         string
	RestaurantID string
	Status       enums.MenuStatus
	Categories   map[string]*menu.Category
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
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
	err = executor.QueryRowContext(ctx, QueryGetMenuIDByUUID, menuAgg.ID().String()).Scan(&menuDbId)
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

	for _, event := range menuAgg.PullEvent() {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		_, err = executor.ExecContext(ctx, QueryInsertOutboxEvent,
			event.EventID().String(),
			menuAgg.ID().String(),
			"Menu",
			event.EventName(),
			payload,
			event.OccurredOn())

		if err != nil {
			return fmt.Errorf("failed to save outbox event: %w", err)
		}
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
		return nil, output.ErrEntityNotFound
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
			menuStatus, _ := enums.ParseMenuStatus(row.MenuStatus)

			baskets[row.MenuUUID] = &menuBasket{
				Name:         row.MenuName,
				RestaurantID: row.RestaurantID,
				Status:       menuStatus,
				Categories:   make(map[string]*menu.Category),
			}
			order = append(order, row.MenuUUID)
		}

		processChildren(row, baskets[row.MenuUUID].Categories)
	}

	result := make([]*menu.Menu, 0, len(baskets))

	for _, uuid := range order {
		b := baskets[uuid]

		finalCats := categoriesToSlice(b.Categories)

		mID, _ := valueobjects.ParseMenuId(uuid)
		rID, _ := valueobjects.ParseRestaurantId(b.RestaurantID)
		menuObj := menu.Restore(mID, b.Name, rID, b.Status, finalCats)
		result = append(result, menuObj)
	}

	return result, nil
}

func processChildren(row dao.MenuCompositeDAO, cats map[string]*menu.Category) {

	if row.CategoryUUID == nil {
		return
	}

	catUUID := *row.CategoryUUID

	if _, exists := cats[catUUID]; !exists {
		cid, _ := valueobjects.ParseCategoryId(catUUID)
		newCat := menu.RestoreCategory(cid, *row.CategoryName, nil)
		cats[catUUID] = newCat
	}

	if row.ItemUUID != nil {
		iid, _ := valueobjects.ParseItemId(*row.ItemUUID)
		price := common.NewMoneyFromCents(*row.ItemPrice)
		status, _ := enums.ParseItemStatus(*row.ItemStatus)

		newItem := menu.RestoreItemMenu(iid, *row.ItemName, *row.ItemDesc, price, status)
		cats[catUUID].AddItem(*newItem)
	}
}

func categoriesToSlice(cats map[string]*menu.Category) []menu.Category {
	result := make([]menu.Category, 0, len(cats))
	for _, c := range cats {
		result = append(result, *c)
	}
	return result
}
