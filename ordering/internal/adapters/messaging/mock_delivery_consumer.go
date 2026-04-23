package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vterry/food-project/common/pkg/domain/event"
)

type MockDeliveryConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewMockDeliveryConsumer(url string) (*MockDeliveryConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare Delivery Exchanges
	_ = ch.ExchangeDeclare("delivery.commands", "topic", true, false, false, false, nil)
	_ = ch.ExchangeDeclare("delivery.events", "topic", true, false, false, false, nil)

	return &MockDeliveryConsumer{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *MockDeliveryConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare("delivery.mock_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	_ = c.channel.QueueBind(q.Name, "delivery.command.schedule", "delivery.commands", false, nil)

	msgs, err := c.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			slog.Info("Mock Delivery received Schedule command")
			
			var payload struct {
				OrderID string `json:"order_id"`
			}
			_ = json.Unmarshal(d.Body, &payload)

			correlationID := ""
			if cid, ok := d.Headers["correlation_id"].(string); ok {
				correlationID = cid
			}

			// Simulate processing time
			slog.Info("Mock Delivery scheduling delivery for order", "order_id", payload.OrderID)
			
			// Respond with DeliveryScheduled event
			respEvent := struct {
				OrderID string `json:"order_id"`
			}{OrderID: payload.OrderID}
			
			envelope := event.EventEnvelope{
				Header: event.Metadata{
					ID:            "ev-mock-del-" + payload.OrderID,
					CorrelationID: correlationID,
					Source:        "delivery-mock",
					Type:          "delivery.scheduled",
					Timestamp:     time.Now(),
				},
				Payload: respEvent,
			}

			body, _ := json.Marshal(envelope)
			_ = c.channel.PublishWithContext(ctx, "delivery.events", "delivery.scheduled", false, false, amqp.Publishing{
				ContentType: "application/json",
				Headers: amqp.Table{
					"correlation_id": correlationID,
				},
				Body: body,
			})

			d.Ack(false)
		}
	}()

	return nil
}

func (c *MockDeliveryConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
