package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"notification-service/internal/domain"
	"notification-service/internal/provider"
)

type Worker struct {
	sender     provider.EmailSender
	maxRetries int
}

func NewWorker(sender provider.EmailSender, maxRetries int) *Worker {
	return &Worker{sender: sender, maxRetries: maxRetries}
}

func (w *Worker) Process(ctx context.Context, event domain.PaymentCompletedEvent) error {
	var lastErr error

	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		if attempt > 1 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("[Worker] Retry %d/%d for event %s — waiting %v",
				attempt, w.maxRetries, event.EventID, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			}
		}

		err := w.sender.Send(ctx, event.CustomerEmail, event.OrderID, event.Amount)
		if err == nil {
			if attempt > 1 {
				log.Printf("[Worker] Success on attempt %d for event %s", attempt, event.EventID)
			}
			return nil
		}

		lastErr = err
		log.Printf("[Worker] Attempt %d failed for event %s: %v", attempt, event.EventID, err)
	}

	return fmt.Errorf("all %d attempts failed, last error: %w", w.maxRetries, lastErr)
}
