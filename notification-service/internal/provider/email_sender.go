package provider

import "context"

type EmailSender interface {
	Send(ctx context.Context, to string, orderID string, amount int64) error
}
