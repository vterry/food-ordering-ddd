package grpc

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/adapters/input/grpc/pb"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

type CatalogGrpcServer struct {
	pb.UnimplementedCatalogServiceServer
	menuService input.CatalogQueryService
}

func NewCatalogGrpcServer(menuService input.CatalogQueryService) *CatalogGrpcServer {
	return &CatalogGrpcServer{
		menuService: menuService,
	}
}

func (c *CatalogGrpcServer) ValidateRestaurantAndItems(ctx context.Context, req *pb.ValidateRestaurantAndItemsRequest) (*pb.ValidateRestaurantAndItemsResponse, error) {
	validationReq := input.ValidateOrderRequest{
		RestaurantID: req.RestaurantId,
		ItemIDs:      req.ItemsId,
	}

	res, err := c.menuService.ValidateOrder(ctx, validationReq)
	if err != nil {
		return nil, err
	}

	pbItems := make([]*pb.ItemSnapshot, 0, len(res.Items))
	for _, item := range res.Items {
		pbItems = append(pbItems, &pb.ItemSnapshot{
			ItemId:     item.ID,
			Name:       item.Name,
			PriceCents: item.PriceCents,
		})
	}
	return &pb.ValidateRestaurantAndItemsResponse{
		Valid:            res.Valid,
		ValidationErrors: res.ValidationErrors,
		Items:            pbItems,
	}, nil
}
