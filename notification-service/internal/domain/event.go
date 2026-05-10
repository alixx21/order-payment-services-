package domain

import "time"

// PaymentCompletedEvent is the domain event consumed from RabbitMQ.
// The Notification Service only knows about this event — it knows nothing
// about Order Service or Payment Service internals.
type PaymentCompletedEvent struct {
	EventID       string
	OrderID       string
	Amount        int64 // in cents
	CustomerEmail string
	Status        string
}

// ProcessedEvent represents a record of an already-handled event.
// Used for idempotency — stored in PostgreSQL.
type ProcessedEvent struct {
	EventID     string
	ProcessedAt time.Time
}
