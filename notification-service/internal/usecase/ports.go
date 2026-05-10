package usecase

import "context"

// IdempotencyRepository checks and records processed events.
// Using PostgreSQL (not in-memory map) so duplicates are detected
// even after service restarts.
type IdempotencyRepository interface {
	// IsProcessed returns true if this event_id was already handled.
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	// MarkProcessed inserts the event_id so future duplicates are skipped.
	MarkProcessed(ctx context.Context, eventID string) error
}
