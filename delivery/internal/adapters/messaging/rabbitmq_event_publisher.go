package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/event"
)

type RabbitMQEventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	source  string
}

func NewRabbitMQEventPublisher(url string, source string) (*RabbitMQEventPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare Exchange
	err = ch.ExchangeDeclare(
		"delivery.events", // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQEventPublisher{
		conn:    conn,
		channel: ch,
		source:  source,
	}, nil
}

func (p *RabbitMQEventPublisher) Publish(ctx context.Context, events ...base.DomainEvent) error {
	correlationID := ctxutil.GetCorrelationID(ctx)

	for _, e := range events {
		slog.Info("Publishing event", "type", e.EventType(), "correlation_id", correlationID)
		
		envelope := event.Wrap(e, correlationID, p.source)
		
		body, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		err = p.channel.PublishWithContext(
			ctx,
			"delivery.events", // exchange
			e.EventType(),     // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Headers: amqp.Table{
					"correlation_id": correlationID,
				},
				Body: body,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
	}

	return nil
}

func (p *RabbitMQEventPublisher) PublishRaw(ctx context.Context, eventType string, correlationID string, payload []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		"delivery.events",
		eventType,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers: amqp.Table{
				"correlation_id": correlationID,
			},
			Body: payload,
		},
	)
}

func (p *RabbitMQEventPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
