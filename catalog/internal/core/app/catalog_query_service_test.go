package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	inputPkg "github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

type mockCatalogQueryRepo struct {
	findOrderValidationDataMock func(ctx context.Context, restaurantID string, itemIDs []string) (*output.OrderValidationData, error)
}

func (m *mockCatalogQueryRepo) FindOrderValidationData(ctx context.Context, restaurantID string, itemIDs []string) (*output.OrderValidationData, error) {
	return m.findOrderValidationDataMock(ctx, restaurantID, itemIDs)
}

func (m *mockCatalogQueryRepo) FindActiveMenuRows(ctx context.Context, restaurantId string) ([]output.ActiveMenuRow, error) {
	return nil, nil // Return empty mock for now, tests not testing this
}

func TestCatalogQueryAppService_ValidateOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		req        inputPkg.ValidateOrderRequest
		setupRepo  func(r *mockCatalogQueryRepo)
		wantValid  bool
		wantErrors []string
		wantItems  []inputPkg.ItemSnapshot
		wantErr    bool
	}{
		{
			name: "repository error returns error",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return nil, output.ErrEntityNotFound
				}
			},
			wantErr: true,
		},
		{
			name: "closed restaurant returns Valid false",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return &output.OrderValidationData{
						RestaurantUUID:   "rest-123",
						RestaurantStatus: enums.RestaurantClosed.String(),
						HasActiveMenu:    true,
					}, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{"restaurant cannot accept orders"},
		},
		{
			name: "open restaurant without active menu returns Valid false",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return &output.OrderValidationData{
						RestaurantUUID:   "rest-123",
						RestaurantStatus: enums.RestaurantOpened.String(),
						HasActiveMenu:    false,
					}, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{"restaurant cannot accept orders"},
		},
		{
			name: "all items valid and available returns Valid true with snapshots",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1", "item-2"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return &output.OrderValidationData{
						RestaurantUUID:   "rest-123",
						RestaurantStatus: enums.RestaurantOpened.String(),
						HasActiveMenu:    true,
						Items: []output.OrderValidationItem{
							{ItemUUID: "item-1", ItemName: "Pizza", PriceCents: 2500, ItemStatus: enums.ItemAvailable.String()},
							{ItemUUID: "item-2", ItemName: "Lasagna", PriceCents: 3000, ItemStatus: enums.ItemAvailable.String()},
						},
					}, nil
				}
			},
			wantValid: true,
			wantItems: []inputPkg.ItemSnapshot{
				{ID: "item-1", Name: "Pizza", PriceCents: 2500},
				{ID: "item-2", Name: "Lasagna", PriceCents: 3000},
			},
		},
		{
			name: "item not found in menu returns Valid false",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1", "item-missing"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return &output.OrderValidationData{
						RestaurantUUID:   "rest-123",
						RestaurantStatus: enums.RestaurantOpened.String(),
						HasActiveMenu:    true,
						Items: []output.OrderValidationItem{
							{ItemUUID: "item-1", ItemName: "Pizza", PriceCents: 2500, ItemStatus: enums.ItemAvailable.String()},
						},
					}, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{fmt.Sprintf("item %s not found in active menu", "item-missing")},
		},
		{
			name: "unavailable item returns Valid false",
			req: inputPkg.ValidateOrderRequest{
				RestaurantID: "rest-123",
				ItemIDs:      []string{"item-1"},
			},
			setupRepo: func(r *mockCatalogQueryRepo) {
				r.findOrderValidationDataMock = func(_ context.Context, _ string, _ []string) (*output.OrderValidationData, error) {
					return &output.OrderValidationData{
						RestaurantUUID:   "rest-123",
						RestaurantStatus: enums.RestaurantOpened.String(),
						HasActiveMenu:    true,
						Items: []output.OrderValidationItem{
							{ItemUUID: "item-1", ItemName: "Pizza", PriceCents: 2500, ItemStatus: enums.ItemTempUnavailable.String()},
						},
					}, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{fmt.Sprintf("item %s is unavailable", "item-1")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCatalogQueryRepo{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewCatalogQueryAppService(repo)
			resp, err := svc.ValidateOrder(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, resp.Valid)

			if tt.wantErrors != nil {
				assert.Equal(t, tt.wantErrors, resp.ValidationErrors)
			}

			if tt.wantItems != nil {
				assert.Equal(t, tt.wantItems, resp.Items)
			}
		})
	}
}
