package external

import (
	"context"
	"fmt"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/ports"
	pb "github.com/vterry/food-project/restaurant/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gRPCRestaurantCatalogClient struct {
	client pb.CatalogServiceClient
}

func NewRestaurantCatalogClient(address string) (ports.RestaurantCatalogPort, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to restaurant service: %v", err)
	}

	return &gRPCRestaurantCatalogClient{
		client: pb.NewCatalogServiceClient(conn),
	}, nil
}

func (c *gRPCRestaurantCatalogClient) GetMenuItem(ctx context.Context, restaurantID, itemID vo.ID) (*ports.RestaurantMenuItem, error) {
	resp, err := c.client.GetMenuItem(ctx, &pb.GetMenuItemRequest{
		RestaurantId: restaurantID.String(),
		ItemId:       itemID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get menu item from restaurant service: %v", err)
	}

	price := vo.NewMoneyFromFloat(resp.Price, resp.Currency)

	return &ports.RestaurantMenuItem{
		ID:           vo.NewID(resp.Id),
		RestaurantID: vo.NewID(resp.RestaurantId),
		Name:         resp.Name,
		Price:        price,
		Available:    resp.Available,
	}, nil
}
