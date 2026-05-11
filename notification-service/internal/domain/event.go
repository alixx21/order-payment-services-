package domain

import "time"

type PaymentCompletedEvent struct {
	EventID       string
	OrderID       string
	Amount        int64 // in cents
	CustomerEmail string
	Status        string
}

type ProcessedEvent struct {
	EventID     string
	ProcessedAt time.Time
}
