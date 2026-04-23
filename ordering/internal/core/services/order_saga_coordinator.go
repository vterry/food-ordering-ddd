package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vterry/food-project/ordering/internal/core/domain/order"
	"github.com/vterry/food-project/ordering/internal/core/domain/saga"
	"github.com/vterry/food-project/ordering/internal/core/ports"
)

type OrderSagaCoordinator struct {
	orderRepo   ports.OrderRepository
	sagaRepo    ports.SagaRepository
	publisher   ports.CommandPublisher
}

func NewOrderSagaCoordinator(
	orderRepo ports.OrderRepository,
	sagaRepo ports.SagaRepository,
	publisher ports.CommandPublisher,
) *OrderSagaCoordinator {
	return &OrderSagaCoordinator{
		orderRepo:   orderRepo,
		sagaRepo:    sagaRepo,
		publisher:   publisher,
	}
}

// Event Handlers for Saga Responses

func (c *OrderSagaCoordinator) HandlePaymentAuthorized(ctx context.Context, orderID string, paymentID string) error {
	slog.Info("Handling PaymentAuthorized", "order_id", orderID, "payment_id", paymentID)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkPaymentAuthorized(); err != nil {
		return err
	}

	s.SetStep("CREATE_TICKET")
	s.AddData("payment_id", paymentID)

	if err := c.saveOrderAndSaga(ctx, o, s); err != nil {
		return err
	}

	// Publish CreateTicket command
	items := make([]map[string]interface{}, 0, len(o.Items()))
	for _, item := range o.Items() {
		items = append(items, map[string]interface{}{
			"product_id": item.MenuItemID().String(),
			"name":       item.Name(),
			"quantity":   item.Quantity(),
		})
	}

	payload := map[string]interface{}{
		"order_id":      o.ID().String(),
		"restaurant_id": o.RestaurantID().String(),
		"items":         items,
	}

	return c.publisher.PublishCommand(ctx, "restaurant.commands", "restaurant.command.create_ticket", payload)
}

func (c *OrderSagaCoordinator) HandlePaymentAuthorizationFailed(ctx context.Context, orderID string, reason string) error {
	slog.Info("Handling PaymentAuthorizationFailed", "order_id", orderID, "reason", reason)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.Reject(reason); err != nil {
		return err
	}

	s.SetStatus(saga.SagaFailed)

	return c.saveOrderAndSaga(ctx, o, s)
}

func (c *OrderSagaCoordinator) HandleTicketConfirmed(ctx context.Context, orderID string) error {
	slog.Info("Handling TicketConfirmed", "order_id", orderID)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkRestaurantConfirmed(); err != nil {
		return err
	}

	s.SetStep("CAPTURE_PAYMENT")

	if err := c.saveOrderAndSaga(ctx, o, s); err != nil {
		return err
	}

	// Publish CapturePayment command
	paymentID := s.Data()["payment_id"].(string)
	payload := map[string]interface{}{
		"payment_id": paymentID,
	}

	return c.publisher.PublishCommand(ctx, "payment.commands", "payment.command.capture", payload)
}

func (c *OrderSagaCoordinator) HandleTicketRejected(ctx context.Context, orderID string, reason string) error {
	slog.Info("Handling TicketRejected", "order_id", orderID, "reason", reason)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkRestaurantRejected(reason); err != nil {
		return err
	}

	s.SetStatus(saga.SagaCompensating)
	s.SetStep("RELEASE_PAYMENT")

	if err := c.saveOrderAndSaga(ctx, o, s); err != nil {
		return err
	}

	// Publish ReleasePayment command
	paymentID := s.Data()["payment_id"].(string)
	payload := map[string]interface{}{
		"payment_id": paymentID,
	}

	return c.publisher.PublishCommand(ctx, "payment.commands", "payment.command.release", payload)
}

func (c *OrderSagaCoordinator) HandlePaymentCaptured(ctx context.Context, orderID string) error {
	slog.Info("Handling PaymentCaptured", "order_id", orderID)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkPaymentCaptured(); err != nil {
		return err
	}

	s.SetStep("SCHEDULE_DELIVERY")

	if err := c.saveOrderAndSaga(ctx, o, s); err != nil {
		return err
	}

	// Publish ScheduleDelivery command
	payload := map[string]interface{}{
		"order_id": orderID,
		// In a real system, we would pass restaurant and customer addresses
	}

	return c.publisher.PublishCommand(ctx, "delivery.commands", "delivery.command.schedule", payload)
}

func (c *OrderSagaCoordinator) HandleDeliveryScheduled(ctx context.Context, orderID string) error {
	slog.Info("Handling DeliveryScheduled", "order_id", orderID)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkDeliveryScheduled(); err != nil {
		return err
	}

	s.SetStatus(saga.SagaCompleted)

	return c.saveOrderAndSaga(ctx, o, s)
}

func (c *OrderSagaCoordinator) HandlePaymentReleased(ctx context.Context, orderID string) error {
	slog.Info("Handling PaymentReleased", "order_id", orderID)
	
	o, s, err := c.loadOrderAndSaga(ctx, orderID)
	if err != nil {
		return err
	}

	if err := o.MarkCancelled(); err != nil {
		return err
	}

	s.SetStatus(saga.SagaFailed)

	return c.saveOrderAndSaga(ctx, o, s)
}

// Helpers

func (c *OrderSagaCoordinator) loadOrderAndSaga(ctx context.Context, orderID string) (*order.Order, *saga.SagaState, error) {
	o, err := c.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load order %s: %w", orderID, err)
	}
	if o == nil {
		return nil, nil, fmt.Errorf("order %s not found", orderID)
	}

	s, err := c.sagaRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load saga for order %s: %w", orderID, err)
	}
	if s == nil {
		return nil, nil, fmt.Errorf("saga for order %s not found", orderID)
	}

	return o, s, nil
}

func (c *OrderSagaCoordinator) saveOrderAndSaga(ctx context.Context, o *order.Order, s *saga.SagaState) error {
	// We should ideally do this in a single transaction, but since repositories are abstracted,
	// we assume the implementation handles it or we accept eventual consistency here if 
	// we use idempotent consumers.
	
	// Implementation note: The persistence layer should handle the version check for 'o'
	if err := c.orderRepo.UpdateWithVersion(ctx, o, o.Version()-1); err != nil {
		return fmt.Errorf("failed to update order with version check: %w", err)
	}

	if err := c.sagaRepo.Save(ctx, s); err != nil {
		return fmt.Errorf("failed to save saga state: %w", err)
	}

	return nil
}
