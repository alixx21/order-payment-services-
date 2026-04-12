package domain

import (
	"errors"
	"time"
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64 //cents
	Status         string
	CreatedAt      time.Time
	IdempotencyKey string
}

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidAmount = errors.New("amount must be greater than 0")
	ErrCannotCancel  = errors.New("only pending orders can be cancelled")
)

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)
