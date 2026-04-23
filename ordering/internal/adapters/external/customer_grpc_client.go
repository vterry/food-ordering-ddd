package external

import (
	"context"
	"fmt"

	pb "github.com/vterry/food-project/customer/api/proto"
	"github.com/vterry/food-project/ordering/internal/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CustomerGRPCClient struct {
	client pb.CustomerServiceClient
}

func NewCustomerGRPCClient(address string) (*CustomerGRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to customer service: %w", err)
	}

	return &CustomerGRPCClient{
		client: pb.NewCustomerServiceClient(conn),
	}, nil
}

func (c *CustomerGRPCClient) GetCustomerByID(ctx context.Context, customerID string) (*ports.CustomerDTO, error) {
	resp, err := c.client.GetCustomerByID(ctx, &pb.GetCustomerRequest{Id: customerID})
	if err != nil {
		return nil, err
	}

	return &ports.CustomerDTO{
		ID:    resp.Id,
		Name:  resp.Name,
		Email: resp.Email,
	}, nil
}
