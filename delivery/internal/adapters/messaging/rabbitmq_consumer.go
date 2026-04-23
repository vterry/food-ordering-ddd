package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vterry/food-project/common/pkg/messaging"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
	"github.com/vterry/food-project/delivery/internal/core/ports"
)

type RabbitMQConsumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	useCase     ports.DeliveryUseCase
	idemHandler *messaging.IdempotentHandler
}

func NewRabbitMQConsumer(url string, useCase ports.DeliveryUseCase, idemHandler *messaging.IdempotentHandler) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQConsumer{
		conn:        conn,
		channel:     ch,
		useCase:     useCase,
		idemHandler: idemHandler,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"delivery.commands", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}

	// Bind commands from ordering exchange
	commands := []string{
		"delivery.command.schedule",
		"delivery.command.cancel",
	}
	for _, rk := range commands {
		err = c.channel.QueueBind(q.Name, rk, "delivery.commands", false, nil)
		if err != nil {
			return err
		}
	}

	msgs, err := c.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			slog.Info("Delivery Command received", "routing_key", d.RoutingKey)

			correlationID := ""
			if cid, ok := d.Headers["correlation_id"].(string); ok {
				correlationID = cid
			}
			// In a real scenario, we'd use the correlationID in the context

			var err error
			msgID := d.MessageId
			if msgID == "" {
				msgID = correlationID // Fallback
			}

			err = c.idemHandler.Handle(ctx, msgID, func(ctx context.Context) error {
				switch d.RoutingKey {
				case "delivery.command.schedule":
					return c.handleScheduleDelivery(ctx, d.Body, correlationID)
				case "delivery.command.cancel":
					return c.handleCancelDelivery(ctx, d.Body)
				default:
					return nil
				}
			})

			if err != nil {
				slog.Error("Error processing delivery command", "error", err, "routing_key", d.RoutingKey)
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) handleScheduleDelivery(ctx context.Context, body []byte, correlationID string) error {
	var cmd struct {
		OrderID      string           `json:"order_id"`
		RestaurantID string           `json:"restaurant_id"`
		CustomerID   string           `json:"customer_id"`
		Address      delivery.Address `json:"address"`
	}
	if err := json.Unmarshal(body, &cmd); err != nil {
		return err
	}

	return c.useCase.ScheduleDelivery(ctx, ports.ScheduleDeliveryCommand{
		OrderID:       vo.NewID(cmd.OrderID),
		RestaurantID:  vo.NewID(cmd.RestaurantID),
		CustomerID:    vo.NewID(cmd.CustomerID),
		Address:       cmd.Address,
		CorrelationID: vo.NewID(correlationID),
	})
}

func (c *RabbitMQConsumer) handleCancelDelivery(ctx context.Context, body []byte) error {
	var cmd struct {
		DeliveryID string `json:"delivery_id"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal(body, &cmd); err != nil {
		return err
	}

	return c.useCase.CancelDelivery(ctx, vo.NewID(cmd.DeliveryID), cmd.Reason)
}

func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
