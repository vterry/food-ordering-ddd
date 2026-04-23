package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQEventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQEventPublisher(url string) (*RabbitMQEventPublisher, error) {
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
		"ordering.events", // name
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
	}, nil
}

func (p *RabbitMQEventPublisher) PublishRaw(ctx context.Context, eventType string, correlationID string, payload []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		"ordering.events", // exchange
		eventType,         // routing key
		false,             // mandatory
		false,             // immediate
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
