package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/event"
)

type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	source  string
}

func NewRabbitMQPublisher(url string, source string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare Exchange for payment events
	err = ch.ExchangeDeclare(
		"payment.events", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
		source:  source,
	}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, events ...base.DomainEvent) error {
	correlationID := ctxutil.GetCorrelationID(ctx)

	for _, e := range events {
		envelope := event.Wrap(e, correlationID, p.source)
		
		body, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		err = p.channel.PublishWithContext(
			ctx,
			"payment.events", // exchange
			e.EventType(),    // routing key
			false,            // mandatory
			false,            // immediate
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

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
