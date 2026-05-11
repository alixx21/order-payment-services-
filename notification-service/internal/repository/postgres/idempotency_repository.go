package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type IdempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1)`,
		eventID,
	).Scan(&exists)
	return exists, err
}

func (r *IdempotencyRepository) MarkProcessed(ctx context.Context, eventID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`,
		eventID,
	)
	return err
}
