package usecase

import (
	"context"
	"fmt"
	"log"

	"notification-service/internal/domain"
	"notification-service/internal/worker"
)

type NotificationUseCase struct {
	repo           IdempotencyRepository
	jobIdempotency JobIdempotency
	worker         *worker.Worker
}

func New(repo IdempotencyRepository, jobIdempotency JobIdempotency, w *worker.Worker) *NotificationUseCase {
	return &NotificationUseCase{
		repo:           repo,
		jobIdempotency: jobIdempotency,
		worker:         w,
	}
}

func (uc *NotificationUseCase) HandlePaymentCompleted(ctx context.Context, event domain.PaymentCompletedEvent) error {
	done, err := uc.jobIdempotency.IsJobDone(ctx, event.EventID)
	if err != nil {
		log.Printf("[UC] Redis idempotency check error: %v — falling back to DB", err)
	} else if done {
		log.Printf("[UC] Event %s already processed (Redis cache) — skipping", event.EventID)
		return nil
	}

	already, err := uc.repo.IsProcessed(ctx, event.EventID)
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}
	if already {
		log.Printf("[UC] Event %s already processed (DB) — skipping", event.EventID)
		_ = uc.jobIdempotency.MarkJobDone(ctx, event.EventID)
		return nil
	}

	if err := uc.repo.MarkProcessed(ctx, event.EventID); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	if err := uc.worker.Process(ctx, event); err != nil {
		log.Printf("[UC] Worker failed for event %s after all retries: %v", event.EventID, err)
		return err
	}

	_ = uc.jobIdempotency.MarkJobDone(ctx, event.EventID)

	return nil
}
