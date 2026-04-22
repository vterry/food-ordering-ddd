package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/payment/internal/core/ports"
)

type AuthorizeCommandPayload struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	Token   string `json:"card_token"`
}

type PaymentCommandPayload struct {
	PaymentID string `json:"payment_id"`
}

type RabbitMQConsumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	paymentService ports.PaymentUseCase
}

func NewRabbitMQConsumer(url string, paymentService ports.PaymentUseCase) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare Exchange for payment commands
	err = ch.ExchangeDeclare(
		"payment.commands", // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQConsumer{
		conn:           conn,
		channel:        ch,
		paymentService: paymentService,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"payment.payment_commands", // name
		true,                       // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		return err
	}

	routingKeys := []string{
		"payment.command.authorize",
		"payment.command.capture",
		"payment.command.refund",
		"payment.command.release",
	}

	for _, rk := range routingKeys {
		err = c.channel.QueueBind(q.Name, rk, "payment.commands", false, nil)
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
			slog.Info("Received a payment command", "routing_key", d.RoutingKey)

			correlationID := ""
			if cid, ok := d.Headers["correlation_id"].(string); ok {
				correlationID = cid
			}

			mctx := ctxutil.WithCorrelationID(ctx, correlationID)

			var err error
			switch d.RoutingKey {
			case "payment.command.authorize":
				err = c.handleAuthorize(mctx, d.Body)
			case "payment.command.capture":
				err = c.handleCapture(mctx, d.Body)
			case "payment.command.refund":
				err = c.handleRefund(mctx, d.Body)
			case "payment.command.release":
				err = c.handleRelease(mctx, d.Body)
			}

			if err != nil {
				slog.Error("Error handling payment command", "error", err, "routing_key", d.RoutingKey)
				d.Nack(false, true) // Requeue
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *RabbitMQConsumer) handleAuthorize(ctx context.Context, body []byte) error {
	var payload AuthorizeCommandPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	_, err := c.paymentService.Authorize(ctx, ports.AuthorizeCommand(payload))
	return err
}

func (c *RabbitMQConsumer) handleCapture(ctx context.Context, body []byte) error {
	var payload PaymentCommandPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return c.paymentService.Capture(ctx, ports.CaptureCommand(payload))
}

func (c *RabbitMQConsumer) handleRefund(ctx context.Context, body []byte) error {
	var payload PaymentCommandPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return c.paymentService.Refund(ctx, ports.RefundCommand(payload))
}

func (c *RabbitMQConsumer) handleRelease(ctx context.Context, body []byte) error {
	var payload PaymentCommandPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return c.paymentService.Release(ctx, ports.ReleaseCommand(payload))
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
