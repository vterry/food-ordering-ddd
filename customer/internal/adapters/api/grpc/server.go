package grpc

import (
	"context"

	pb "github.com/vterry/food-project/customer/api/proto"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomerGRPCServer struct {
	pb.UnimplementedCustomerServiceServer
	service *services.CustomerService
}

func NewCustomerGRPCServer(service *services.CustomerService) *CustomerGRPCServer {
	return &CustomerGRPCServer{
		service: service,
	}
}

func (s *CustomerGRPCServer) GetCustomerByID(ctx context.Context, req *pb.GetCustomerRequest) (*pb.CustomerResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	cust, err := s.service.GetCustomer(ctx, vo.NewID(req.GetId()))
	if err != nil {
		// Basic mapping of errors
		if err == services.ErrCustomerNotFound {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	addresses := make([]*pb.Address, 0, len(cust.Addresses()))
	for _, addr := range cust.Addresses() {
		addresses = append(addresses, &pb.Address{
			Id:        addr.ID().String(),
			Street:    addr.Street(),
			City:      addr.City(),
			ZipCode:   addr.ZipCode(),
			IsDefault: addr.IsDefault(),
		})
	}

	return &pb.CustomerResponse{
		Id:        cust.ID().String(),
		Name:      cust.Name().String(),
		Email:     cust.Email().String(),
		Phone:     cust.Phone().String(),
		Addresses: addresses,
	}, nil
}
