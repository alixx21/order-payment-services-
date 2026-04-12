package postgres

import (
	"context"
	"database/sql"

	"order-service/internal/domain"

	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt, nullableString(o.IdempotencyKey),
	)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, customer_id, item_name, amount, status, created_at, COALESCE(idempotency_key, '')
		FROM orders WHERE id = $1`, id)

	return scanOrder(row)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, customer_id, item_name, amount, status, created_at, COALESCE(idempotency_key, '')FROM orders WHERE idempotency_key = $1`, key)

	return scanOrder(row)
}

func scanOrder(row *sql.Row) (*domain.Order, error) {
	var o domain.Order
	err := row.Scan(
		&o.ID, &o.CustomerID, &o.ItemName,
		&o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *OrderRepository) GetRevenueByCustomer(ctx context.Context, customerID string) (totalAmount int64, ordersCount int, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount),0),COUNT(*) FROM orders WHERE customer_id = $1 AND status = 'Paid'`, customerID)

	err = row.Scan(&totalAmount, &ordersCount)
	return

}
