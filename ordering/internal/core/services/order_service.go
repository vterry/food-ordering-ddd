package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/ordering/internal/core/domain/order"
	"github.com/vterry/food-project/ordering/internal/core/domain/saga"
	"github.com/vterry/food-project/ordering/internal/core/ports"
)

type OrderService struct {
	orderRepo    ports.OrderRepository
	sagaRepo     ports.SagaRepository
	publisher    ports.CommandPublisher
	customerSvc  ports.CustomerClient
}

func NewOrderService(
	orderRepo ports.OrderRepository,
	sagaRepo ports.SagaRepository,
	publisher ports.CommandPublisher,
	customerSvc ports.CustomerClient,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		sagaRepo:     sagaRepo,
		publisher:    publisher,
		customerSvc:  customerSvc,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, cmd ports.CreateOrderCommand) (string, error) {
	slog.Info("Creating order", "customer_id", cmd.CustomerID, "restaurant_id", cmd.RestaurantID)

	// 1. Validate Customer via gRPC
	customer, err := s.customerSvc.GetCustomerByID(ctx, cmd.CustomerID)
	if err != nil {
		return "", fmt.Errorf("failed to validate customer: %w", err)
	}
	if customer == nil {
		return "", fmt.Errorf("customer not found")
	}

	// 2. Map Items
	orderItems := make([]order.OrderItem, 0, len(cmd.Items))
	for _, it := range cmd.Items {
		item := order.NewOrderItem(
			vo.NewID(it.ProductID),
			it.Name,
			it.Quantity,
			vo.NewMoneyFromFloat(it.Price, "BRL"),
			"",
		)
		orderItems = append(orderItems, item)
	}

	// 3. Create Order Aggregate
	orderID := vo.NewID(uuid.New().String())
	o, err := order.NewOrder(
		orderID,
		vo.NewID(cmd.CustomerID),
		vo.NewID(cmd.RestaurantID),
		orderItems,
		"Customer Address",
		vo.NewID(cmd.CorrelationID),
	)
	if err != nil {
		return "", err
	}

	// 4. Update order state
	if err := o.MarkPaymentPending(); err != nil {
		return "", err
	}

	// 5. Create Saga State
	sagaState := saga.NewSagaState(o.ID().String(), "AUTHORIZE_PAYMENT")
	sagaState.AddData("card_token", cmd.CardToken)

	// 6. Persist
	if err := s.orderRepo.Save(ctx, o); err != nil {
		return "", err
	}
	if err := s.sagaRepo.Save(ctx, sagaState); err != nil {
		return "", err
	}

	// 6. Start Saga (Publish first command)
	err = s.publisher.PublishCommand(ctx, "payment.commands", "payment.command.authorize", map[string]interface{}{
		"order_id":   o.ID().String(),
		"amount":     int64(o.TotalAmount().Amount() * 100), // to cents
		"card_token": cmd.CardToken,
	})
	if err != nil {
		slog.Error("Failed to publish initial saga command", "error", err, "order_id", o.ID().String())
	}

	return o.ID().String(), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string) error {
	slog.Info("Cancelling order", "order_id", orderID)

	o, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return fmt.Errorf("order not found")
	}

	if err := o.Cancel(); err != nil {
		return err
	}

	sagaState, err := s.sagaRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	
	if sagaState != nil {
		sagaState.SetStatus(saga.SagaCompensating)
		sagaState.SetStep("CANCEL_COMPENSATION")
		if err := s.sagaRepo.Save(ctx, sagaState); err != nil {
			return err
		}
	}

	if err := s.orderRepo.UpdateWithVersion(ctx, o, o.Version()-1); err != nil {
		return err
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	return s.orderRepo.FindByID(ctx, orderID)
}
