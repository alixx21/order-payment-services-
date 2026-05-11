package usecase

import "context"

type IdempotencyRepository interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type JobIdempotency interface {
	IsJobDone(ctx context.Context, eventID string) (bool, error)
	MarkJobDone(ctx context.Context, eventID string) error
}
