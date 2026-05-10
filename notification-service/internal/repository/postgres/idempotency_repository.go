package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

// IdempotencyRepository stores processed event IDs in PostgreSQL.
// Using DB (not in-memory) means duplicates are detected even after restart.
type IdempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

// IsProcessed checks whether event_id exists in processed_events table.
func (r *IdempotencyRepository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1)`,
		eventID,
	).Scan(&exists)
	return exists, err
}

// MarkProcessed inserts the event_id into processed_events.
// If a duplicate insert happens (race condition), it is safely ignored.
func (r *IdempotencyRepository) MarkProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`,
		eventID,
	)
	return err
}
