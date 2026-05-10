package usecase

import (
	"context"
	"fmt"
	"log"

	"notification-service/internal/domain"
)

// NotificationUseCase handles the business logic for processing payment events.
// It checks idempotency and simulates sending an email notification.
type NotificationUseCase struct {
	repo IdempotencyRepository
}

func New(repo IdempotencyRepository) *NotificationUseCase {
	return &NotificationUseCase{repo: repo}
}

// HandlePaymentCompleted processes a payment.completed event.
//
// Flow:
//  1. Check if event_id was already processed (idempotency).
//  2. If duplicate → skip silently, return nil (consumer will ACK safely).
//  3. If new → mark as processed in DB → simulate sending email.
//
// Returns error only for system failures (DB down, etc.) — consumer will NACK.
func (uc *NotificationUseCase) HandlePaymentCompleted(ctx context.Context, event domain.PaymentCompletedEvent) error {
	// --- Idempotency check ---
	already, err := uc.repo.IsProcessed(ctx, event.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if already {
		log.Printf("[Notification] Duplicate event %s for order %s — skipping", event.EventID, event.OrderID)
		return nil // ACK duplicate safely — don't process again
	}

	// --- Mark as processed BEFORE logging (prevents double-processing on crash) ---
	if err := uc.repo.MarkProcessed(ctx, event.EventID); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	// --- Simulate sending email ---
	amountDollars := float64(event.Amount) / 100.0
	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
		event.CustomerEmail, event.OrderID, amountDollars)

	return nil
}
