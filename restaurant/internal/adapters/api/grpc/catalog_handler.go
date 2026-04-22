package grpc

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	pb "github.com/vterry/food-project/restaurant/api/proto"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogGRPCServer struct {
	pb.UnimplementedCatalogServiceServer
	queryService ports.RestaurantQueryUseCase
}

func NewCatalogGRPCServer(queryService ports.RestaurantQueryUseCase) *CatalogGRPCServer {
	return &CatalogGRPCServer{
		queryService: queryService,
	}
}

func (s *CatalogGRPCServer) GetMenuItem(ctx context.Context, req *pb.GetMenuItemRequest) (*pb.MenuItemResponse, error) {
	restID := vo.NewID(req.RestaurantId)
	itemID := vo.NewID(req.ItemId)

	info, err := s.queryService.GetMenuItemInfo(ctx, restID, itemID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "menu item not found: %v", err)
	}

	return &pb.MenuItemResponse{
		Id:             info.ID.String(),
		RestaurantId:   info.RestaurantID.String(),
		RestaurantName: info.RestaurantName,
		Name:           info.Name,
		Description:    info.Description,
		Price:          info.Price.Amount(),
		Currency:       info.Price.Currency(),
		Available:      info.IsAvailable,
	}, nil
}

func (s *CatalogGRPCServer) ListMenuItems(ctx context.Context, req *pb.ListMenuItemsRequest) (*pb.ListMenuItemsResponse, error) {
	restID := vo.NewID(req.RestaurantId)

	infos, err := s.queryService.ListRestaurantMenuItems(ctx, restID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list menu items: %v", err)
	}

	items := make([]*pb.MenuItemResponse, 0, len(infos))
	for _, info := range infos {
		items = append(items, &pb.MenuItemResponse{
			Id:             info.ID.String(),
			RestaurantId:   info.RestaurantID.String(),
			RestaurantName: info.RestaurantName,
			Name:           info.Name,
			Description:    info.Description,
			Price:          info.Price.Amount(),
			Currency:       info.Price.Currency(),
			Available:      info.IsAvailable,
		})
	}

	return &pb.ListMenuItemsResponse{
		Items: items,
	}, nil
}
