package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

type RabbitMQPublisher struct {
	channel      *amqp.Channel
	exchangeName string
	logger       *slog.Logger
}

func NewRabbitMQPublisher(channel *amqp.Channel, exchangeName string, logger *slog.Logger) (*RabbitMQPublisher, error) {
	err := channel.ExchangeDeclare(
		exchangeName, // Name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange %s error: %w", exchangeName, err)
	}

	return &RabbitMQPublisher{
		channel:      channel,
		exchangeName: exchangeName,
		logger:       logger.With("component", "RabbitMQPublisher"),
	}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, message output.EventMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	routingKey := fmt.Sprintf("%s.%s.%s", message.AggregateType, message.AggregateID, message.EventType)

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(pubCtx,
		p.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    message.EventID,
			Timestamp:    time.UnixMilli(message.OcurredAt),
			Body:         body,
		})

	if err != nil {
		p.logger.Error("failed to publish to rabbitmq", "event_id", message.EventID, "routing_key", routingKey, "error", err)
		return fmt.Errorf("failed to publish to rabbitmq: %w", err)
	}

	p.logger.Debug("event published to rabbitmq successfully", "event_id", message.EventID, "routing_key", routingKey)
	return nil
}
