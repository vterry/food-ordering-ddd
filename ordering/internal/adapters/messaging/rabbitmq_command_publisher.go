package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
)

type RabbitMQCommandPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQCommandPublisher(url string) (*RabbitMQCommandPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Exchanges for commands
	commandExchanges := []string{"payment.commands", "restaurant.commands", "delivery.commands"}
	for _, ex := range commandExchanges {
		err = ch.ExchangeDeclare(
			ex,      // name
			"topic", // type
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare exchange %s: %w", ex, err)
		}
	}

	return &RabbitMQCommandPublisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *RabbitMQCommandPublisher) PublishCommand(ctx context.Context, exchange, routingKey string, payload interface{}) error {
	correlationID := ctxutil.GetCorrelationID(ctx)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal command payload: %w", err)
	}

	return p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Headers: amqp.Table{
				"correlation_id": correlationID,
			},
			Body: body,
		},
	)
}

func (p *RabbitMQCommandPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
