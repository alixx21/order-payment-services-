package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"notification-service/internal/domain"
	"notification-service/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "payment.completed"

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	uc      *usecase.NotificationUseCase
}

func NewConsumer(url string, uc *usecase.NotificationUseCase) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Printf("[RabbitMQ] Consumer connected, listening on queue '%s'", queueName)
	return &Consumer{conn: conn, channel: ch, uc: uc}, nil
}

type messagePayload struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		queueName, "notification-consumer",
		false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	log.Println("[RabbitMQ] Waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[RabbitMQ] Context cancelled, stopping consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var payload messagePayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Printf("[RabbitMQ] Failed to parse message: %v — NACK", err)
		msg.Nack(false, false)
		return
	}

	event := domain.PaymentCompletedEvent{
		EventID:       payload.EventID,
		OrderID:       payload.OrderID,
		Amount:        payload.Amount,
		CustomerEmail: payload.CustomerEmail,
		Status:        payload.Status,
	}

	if err := c.uc.HandlePaymentCompleted(ctx, event); err != nil {
		log.Printf("[RabbitMQ] Processing failed for event %s: %v — NACK", payload.EventID, err)
		msg.Nack(false, false)
		return
	}

	msg.Ack(false)
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
