package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-ordering/catalog/internal/adapters/input/grpc/pb"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

// MockMenuService implements input.MenuService for testing purposes
type mockMenuService struct {
	validateOrderFunc func(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error)
}

// Ensure mockMenuService adheres to the input.MenuService interface
// We only need ValidateOrder for this test, but Go requires all methods.
// We'll panic on unused methods to ensure test isolation if they are accidentally called.
func (m *mockMenuService) ActiveMenu(ctx context.Context, menuIdstr string) error {
	panic("not implemented")
}

func (m *mockMenuService) AddCategory(ctx context.Context, menuIdstr string, req input.AddCategoryRequest) error {
	panic("not implemented")
}

func (m *mockMenuService) AddItemToCategory(ctx context.Context, menuIdstr string, categoryIdstr string, req input.AddItemRequest) error {
	panic("not implemented")
}

func (m *mockMenuService) ArchiveMenu(ctx context.Context, menuIdstr string) error {
	panic("not implemented")
}

func (m *mockMenuService) CreateMenu(ctx context.Context, restIdstr string, req input.CreateMenuRequest) (*input.MenuResponse, error) {
	panic("not implemented")
}

func (m *mockMenuService) GetActiveMenu(ctx context.Context, restaurantIdStr string) (*input.MenuResponse, error) {
	panic("not implemented")
}

func (m *mockMenuService) GetMenu(ctx context.Context, menuIdstr string) (*input.MenuResponse, error) {
	panic("not implemented")
}

func (m *mockMenuService) UpdateItem(ctx context.Context, menuIdStr string, categoryIdStr string, itemIdStr string, req input.UpdateItemRequest) error {
	panic("not implemented")
}

func (m *mockMenuService) ValidateOrder(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
	if m.validateOrderFunc != nil {
		return m.validateOrderFunc(ctx, req)
	}
	return nil, errors.New("validateOrderFunc not implemented in mock")
}

func TestCatalogGrpcServer_ValidateRestaurantAndItems(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.ValidateRestaurantAndItemsRequest
		mockBehavior   func(m *mockMenuService)
		expectedResp   *pb.ValidateRestauranteAndItemsRespose
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name: "Success - All items valid",
			req: &pb.ValidateRestaurantAndItemsRequest{
				RestaurantId: "rest-123",
				ItemsId:      []string{"item-1", "item-2"},
			},
			mockBehavior: func(m *mockMenuService) {
				m.validateOrderFunc = func(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
					assert.Equal(t, "rest-123", req.RestaurantID)
					assert.Equal(t, []string{"item-1", "item-2"}, req.ItemIDs)
					return &input.ValidateOrderResponse{
						Valid: true,
						Items: []input.ItemSnapshot{
							{ID: "item-1", Name: "Item 1", PriceCents: 1000},
							{ID: "item-2", Name: "Item 2", PriceCents: 2000},
						},
						ValidationErrors: []string{},
					}, nil
				}
			},
			expectedResp: &pb.ValidateRestauranteAndItemsRespose{
				Valid: true,
				Items: []*pb.ItemSnapshot{
					{ItemId: "item-1", Name: "Item 1", PriceCents: 1000},
					{ItemId: "item-2", Name: "Item 2", PriceCents: 2000},
				},
				ValidationErrors: []string{},
			},
			expectedErr: false,
		},
		{
			name: "Success - Business Validation Failed (Restaurant Closed)",
			req: &pb.ValidateRestaurantAndItemsRequest{
				RestaurantId: "rest-closed",
				ItemsId:      []string{"item-1"},
			},
			mockBehavior: func(m *mockMenuService) {
				m.validateOrderFunc = func(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
					return &input.ValidateOrderResponse{
						Valid:            false,
						ValidationErrors: []string{"Restaurant is closed"},
						Items:            []input.ItemSnapshot{},
					}, nil
				}
			},
			expectedResp: &pb.ValidateRestauranteAndItemsRespose{
				Valid:            false,
				Items:            []*pb.ItemSnapshot{},
				ValidationErrors: []string{"Restaurant is closed"},
			},
			expectedErr: false,
		},
		{
			name: "Error - Internal Service Error",
			req: &pb.ValidateRestaurantAndItemsRequest{
				RestaurantId: "rest-error",
				ItemsId:      []string{"item-1"},
			},
			mockBehavior: func(m *mockMenuService) {
				m.validateOrderFunc = func(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
					return nil, errors.New("db connection failed")
				}
			},
			expectedResp: nil,
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockMenuService{}
			if tt.mockBehavior != nil {
				tt.mockBehavior(mockSvc)
			}

			server := NewCatalogGrpcServer(mockSvc)
			resp, err := server.ValidateRestaurantAndItems(context.Background(), tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Manual comparison for protobuf structs since reflect.DeepEqual can be tricky with generated code
				// But assert.Equal usually works well for simple proto structs
				assert.Equal(t, tt.expectedResp.Valid, resp.Valid)
				assert.Equal(t, len(tt.expectedResp.Items), len(resp.Items))
				assert.Equal(t, tt.expectedResp.ValidationErrors, resp.ValidationErrors)

				if len(tt.expectedResp.Items) == len(resp.Items) {
					for i, expectedItem := range tt.expectedResp.Items {
						assert.Equal(t, expectedItem.ItemId, resp.Items[i].ItemId)
						assert.Equal(t, expectedItem.Name, resp.Items[i].Name)
						assert.Equal(t, expectedItem.PriceCents, resp.Items[i].PriceCents)
					}
				}
			}
		})
	}
}
