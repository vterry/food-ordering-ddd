package external

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/ports"
	pb "github.com/vterry/food-project/restaurant/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gRPCRestaurantCatalogClient struct {
	client pb.CatalogServiceClient
	cb     *gobreaker.CircuitBreaker
}

func NewRestaurantCatalogClient(address string) (ports.RestaurantCatalogPort, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to restaurant service: %v", err)
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "RestaurantCatalog",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	})

	return &gRPCRestaurantCatalogClient{
		client: pb.NewCatalogServiceClient(conn),
		cb:     cb,
	}, nil
}

func (c *gRPCRestaurantCatalogClient) GetMenuItem(ctx context.Context, restaurantID, itemID vo.ID) (*ports.RestaurantMenuItem, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		resp, err := c.client.GetMenuItem(ctx, &pb.GetMenuItemRequest{
			RestaurantId: restaurantID.String(),
			ItemId:       itemID.String(),
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	if err != nil {
		return nil, fmt.Errorf("restaurant service call failed: %w", err)
	}

	resp := result.(*pb.MenuItemResponse)
	price := vo.NewMoneyFromFloat(resp.Price, resp.Currency)

	return &ports.RestaurantMenuItem{
		ID:           vo.NewID(resp.Id),
		RestaurantID: vo.NewID(resp.RestaurantId),
		Name:         resp.Name,
		Price:        price,
		Available:    resp.Available,
	}, nil
}
