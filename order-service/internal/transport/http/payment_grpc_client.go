package transport

import (
	"context"

	pb "order-service/internal/pb"
	"order-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentGRPCClient(client pb.PaymentServiceClient) *PaymentGRPCClient {
	return &PaymentGRPCClient{client: client}
}
func (c *PaymentGRPCClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (string, error) {
	resp, err := c.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			return "", usecase.ErrPaymentServiceUnavailable
		default:
			return "", usecase.ErrPaymentServiceUnavailable
		}
	}

	if resp.Status == "Declined" {
		return "", usecase.ErrPaymentDeclined
	}

	return resp.TransactionId, nil
}
