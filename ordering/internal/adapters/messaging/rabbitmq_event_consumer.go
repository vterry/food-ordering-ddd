package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/ordering/internal/core/services"
)

type RabbitMQEventConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	coordinator *services.OrderSagaCoordinator
}

func NewRabbitMQEventConsumer(url string, coordinator *services.OrderSagaCoordinator) (*RabbitMQEventConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQEventConsumer{
		conn:        conn,
		channel:     ch,
		coordinator: coordinator,
	}, nil
}

func (c *RabbitMQEventConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"ordering.saga_events", // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return err
	}

	// Declare Exchanges before binding
	exchanges := []string{"payment.events", "restaurant.events", "delivery.events"}
	for _, ex := range exchanges {
		err := c.channel.ExchangeDeclare(ex, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex, err)
		}
	}

	// Bindings for Payment Events
	paymentEvents := []string{
		"payment.authorized",
		"payment.authorization_failed",
		"payment.captured",
		"payment.capture_failed",
		"payment.released",
		"payment.refunded",
	}
	for _, rk := range paymentEvents {
		_ = c.channel.QueueBind(q.Name, rk, "payment.events", false, nil)
	}

	// Bindings for Restaurant Events
	restaurantEvents := []string{
		"restaurant.ticket.confirmed",
		"restaurant.ticket.rejected",
		"restaurant.ticket.ready",
	}
	for _, rk := range restaurantEvents {
		_ = c.channel.QueueBind(q.Name, rk, "restaurant.events", false, nil)
	}

	// Bindings for Delivery Events
	deliveryEvents := []string{
		"delivery.scheduled",
		"delivery.refused",
	}
	for _, rk := range deliveryEvents {
		_ = c.channel.QueueBind(q.Name, rk, "delivery.events", false, nil)
	}

	msgs, err := c.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			slog.Info("Saga Event received", "routing_key", d.RoutingKey)

			correlationID := ""
			if cid, ok := d.Headers["correlation_id"].(string); ok {
				correlationID = cid
			}
			mctx := ctxutil.WithCorrelationID(ctx, correlationID)

			var err error
			switch d.RoutingKey {
			case "payment.authorized":
				err = c.handlePaymentAuthorized(mctx, d.Body)
			case "payment.authorization_failed":
				err = c.handlePaymentAuthorizationFailed(mctx, d.Body)
			case "restaurant.ticket.confirmed":
				err = c.handleTicketConfirmed(mctx, d.Body)
			case "restaurant.ticket.rejected":
				err = c.handleTicketRejected(mctx, d.Body)
			case "payment.captured":
				err = c.handlePaymentCaptured(mctx, d.Body)
			case "delivery.scheduled":
				err = c.handleDeliveryScheduled(mctx, d.Body)
			case "payment.released":
				err = c.handlePaymentReleased(mctx, d.Body)
			}

			if err != nil {
				slog.Error("Error processing saga event", "error", err, "routing_key", d.RoutingKey)
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *RabbitMQEventConsumer) handlePaymentAuthorized(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID   string `json:"order_id"`
			PaymentID string `json:"payment_id"`
		} `json:"payload"`
		OrderID   string `json:"order_id"`
		PaymentID string `json:"payment_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	
	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}
	paymentID := env.Payload.PaymentID
	if paymentID == "" {
		paymentID = env.PaymentID
	}

	return c.coordinator.HandlePaymentAuthorized(ctx, orderID, paymentID)
}

func (c *RabbitMQEventConsumer) handlePaymentAuthorizationFailed(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
			Reason  string `json:"reason"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}
	reason := env.Payload.Reason
	if reason == "" {
		reason = env.Reason
	}

	return c.coordinator.HandlePaymentAuthorizationFailed(ctx, orderID, reason)
}

func (c *RabbitMQEventConsumer) handleTicketConfirmed(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}

	return c.coordinator.HandleTicketConfirmed(ctx, orderID)
}

func (c *RabbitMQEventConsumer) handleTicketRejected(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
			Reason  string `json:"reason"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}
	reason := env.Payload.Reason
	if reason == "" {
		reason = env.Reason
	}

	return c.coordinator.HandleTicketRejected(ctx, orderID, reason)
}

func (c *RabbitMQEventConsumer) handlePaymentCaptured(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}

	return c.coordinator.HandlePaymentCaptured(ctx, orderID)
}

func (c *RabbitMQEventConsumer) handleDeliveryScheduled(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}

	return c.coordinator.HandleDeliveryScheduled(ctx, orderID)
}

func (c *RabbitMQEventConsumer) handlePaymentReleased(ctx context.Context, body []byte) error {
	var env struct {
		Payload struct {
			OrderID string `json:"order_id"`
		} `json:"payload"`
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}

	orderID := env.Payload.OrderID
	if orderID == "" {
		orderID = env.OrderID
	}

	return c.coordinator.HandlePaymentReleased(ctx, orderID)
}

func (c *RabbitMQEventConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
