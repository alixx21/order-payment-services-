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

// Consumer listens to RabbitMQ and delegates to the use case.
// It handles manual ACK/NACK — message is only ACKed after successful processing.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	uc      *usecase.NotificationUseCase
}

// NewConsumer connects to RabbitMQ and declares the queue.
func NewConsumer(url string, uc *usecase.NotificationUseCase) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declare durable queue — must match the producer's declaration
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	// Process one message at a time (fair dispatch)
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Printf("[RabbitMQ] Consumer connected, listening on queue '%s'", queueName)
	return &Consumer{conn: conn, channel: ch, uc: uc}, nil
}

// messagePayload mirrors the JSON structure published by Payment Service.
type messagePayload struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

// Start begins consuming messages. It blocks until ctx is cancelled.
// Manual ACK design:
//   - Parse JSON successfully  → continue
//   - Use case processes event → ACK
//   - Any failure              → NACK (requeue=false to avoid infinite loop)
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		queueName,
		"notification-consumer", // consumer tag
		false, // auto-ack = false (MANUAL ACK)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
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
				log.Println("[RabbitMQ] Channel closed")
				return nil
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	// Step 1: Parse JSON
	var payload messagePayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Printf("[RabbitMQ] Failed to parse message: %v — NACK", err)
		msg.Nack(false, false) // requeue=false — bad message, drop it
		return
	}

	// Step 2: Build domain event
	event := domain.PaymentCompletedEvent{
		EventID:       payload.EventID,
		OrderID:       payload.OrderID,
		Amount:        payload.Amount,
		CustomerEmail: payload.CustomerEmail,
		Status:        payload.Status,
	}

	// Step 3: Process via use case (idempotency check + log)
	if err := c.uc.HandlePaymentCompleted(ctx, event); err != nil {
		log.Printf("[RabbitMQ] Processing failed for event %s: %v — NACK", payload.EventID, err)
		msg.Nack(false, false) // requeue=false — avoid infinite retry loop
		return
	}

	// Step 4: ACK only after successful processing
	msg.Ack(false)
}

// Close cleans up connections on graceful shutdown.
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
