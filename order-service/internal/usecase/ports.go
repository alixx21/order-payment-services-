package usecase

import (
	"context"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
	GetRevenueByCustomer(ctx context.Context, customerID string) (totalAmount int64, ordersCount int, err error)
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (transactionID string, err error)
}
