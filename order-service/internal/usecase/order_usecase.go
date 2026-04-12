package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

var ErrPaymentServiceUnavailable = errors.New("payment service unavailable")

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
}

func New(repo OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient}
}

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

type Revenue struct {
	CustomerID  string
	TotalAmount int64
	OrderCount  int
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	if input.IdempotencyKey != "" {
		if existing, err := uc.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
			return existing, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     input.CustomerID,
		ItemName:       input.ItemName,
		Amount:         input.Amount,
		Status:         domain.StatusPending,
		CreatedAt:      time.Now().UTC(),
		IdempotencyKey: input.IdempotencyKey,
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	_, err := uc.paymentClient.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed)
		order.Status = domain.StatusFailed

		if errors.Is(err, ErrPaymentServiceUnavailable) {
			return nil, ErrPaymentServiceUnavailable
		}
		return order, nil
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, domain.StatusPaid); err != nil {
		return nil, fmt.Errorf("update status to paid: %w", err)
	}
	order.Status = domain.StatusPaid
	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.StatusPending {
		return nil, domain.ErrCannotCancel
	}
	if err := uc.repo.UpdateStatus(ctx, id, domain.StatusCancelled); err != nil {
		return nil, fmt.Errorf("update status to cancelled: %w", err)
	}
	order.Status = domain.StatusCancelled
	return order, nil
}

func (uc *OrderUseCase) GetCustomerRevenue(ctx context.Context, customerID string) (*Revenue, error) {
	if customerID == "" {
		return nil, domain.ErrOrderNotFound
	}
	totalAmount, orderCount, err := uc.repo.GetRevenueByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if orderCount == 0 {
		return nil, domain.ErrOrderNotFound
	}

	return &Revenue{
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		OrderCount:  orderCount,
	}, nil
}
