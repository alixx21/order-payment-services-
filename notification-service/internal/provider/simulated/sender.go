package simulated

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type SimulatedEmailSender struct{}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{}
}

func (s *SimulatedEmailSender) Send(ctx context.Context, to string, orderID string, amount int64) error {
	latency := time.Duration(100+rand.Intn(400)) * time.Millisecond
	select {
	case <-time.After(latency):
	case <-ctx.Done():
		return ctx.Err()
	}

	if rand.Intn(5) == 0 {
		return errors.New("simulated provider error: connection timeout")
	}

	amountDollars := float64(amount) / 100.0
	log.Printf("[Email] Sent to %s for Order #%s. Amount: $%.2f", to, orderID, amountDollars)
	fmt.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f\n", to, orderID, amountDollars)
	return nil
}
