package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	ctxutil "github.com/vterry/food-project/common/pkg/context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

type CreateTicketCommandPayload struct {
	OrderID      string `json:"order_id"`
	RestaurantID string `json:"restaurant_id"`
	Items        []struct {
		ProductID string `json:"product_id"`
		Name      string `json:"name"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
}

type RabbitMQConsumer struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	ticketService ports.TicketUseCase
}

func NewRabbitMQConsumer(url string, ticketService ports.TicketUseCase) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare Exchange for restaurant commands
	err = ch.ExchangeDeclare(
		"restaurant.commands", // name
		"topic",               // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &RabbitMQConsumer{
		conn:          conn,
		channel:       ch,
		ticketService: ticketService,
	}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		"restaurant.ticket_commands", // name
		true,                         // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		q.Name,
		"restaurant.command.create_ticket", // routing key
		"restaurant.commands",             // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name,
		"",    // consumer
		false, // auto-ack (using manual ack for safety)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a command: %s", d.RoutingKey)
			
			correlationID := ""
			if cid, ok := d.Headers["correlation_id"].(string); ok {
				correlationID = cid
			}

			// Injetar Correlation ID no contexto
			mctx := ctxutil.WithCorrelationID(ctx, correlationID)

			switch d.RoutingKey {
			case "restaurant.command.create_ticket":
				if err := c.handleCreateTicket(mctx, d.Body); err != nil {
					log.Printf("Error handling CreateTicket: %v", err)
					d.Nack(false, true) // Requeue
				} else {
					d.Ack(false)
				}
			default:
				log.Printf("Unknown routing key: %s", d.RoutingKey)
				d.Ack(false)
			}
		}
	}()

	log.Println("RabbitMQ Consumer started")
	return nil
}

func (c *RabbitMQConsumer) handleCreateTicket(ctx context.Context, body []byte) error {
	var payload CreateTicketCommandPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	items := make([]ticket.TicketItem, 0, len(payload.Items))
	for _, it := range payload.Items {
		items = append(items, ticket.TicketItem{
			ProductID: vo.NewID(it.ProductID),
			Name:      it.Name,
			Quantity:  it.Quantity,
		})
	}

	cmd := ports.CreateTicketCommand{
		OrderID:      vo.NewID(payload.OrderID),
		RestaurantID: vo.NewID(payload.RestaurantID),
		Items:        items,
	}

	_, err := c.ticketService.CreateTicket(ctx, cmd)
	return err
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
