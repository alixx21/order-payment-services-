package grpc

import (
	"log"
	"time"

	"order-service/internal/usecase"

	orderpb "github.com/alixx21/ap2-generated/order"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCServer struct {
	orderpb.UnimplementedOrderServiceServer
	uc *usecase.OrderUseCase
}

func NewOrderGRPCServer(uc *usecase.OrderUseCase) *OrderGRPCServer {
	return &OrderGRPCServer{uc: uc}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(
	req *orderpb.OrderRequest,
	stream orderpb.OrderService_SubscribeToOrderUpdatesServer,
) error {
	orderID := req.OrderId
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	_, err := s.uc.GetOrder(stream.Context(), orderID)
	if err != nil {
		return status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	var lastStatus string
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Printf("[Stream] Started tracking order: %s", orderID)

	for {
		select {

		case <-stream.Context().Done():
			log.Printf("[Stream] Client disconnected for order: %s", orderID)
			return nil

		case <-ticker.C:
			order, err := s.uc.GetOrder(stream.Context(), orderID)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to read order: %v", err)
			}
			if order.Status != lastStatus {
				lastStatus = order.Status

				update := &orderpb.OrderStatusUpdate{
					OrderId:   order.ID,
					Status:    order.Status,
					UpdatedAt: timestamppb.Now(),
				}

				if err := stream.Send(update); err != nil {
					log.Printf("[Stream] Failed to send update for order %s: %v", orderID, err)
					return err
				}
				log.Printf("[Stream] Pushed update → order %s: %s", orderID, order.Status)
			}
			if order.Status == "Paid" || order.Status == "Failed" || order.Status == "Cancelled" {
				log.Printf("[Stream] Order %s reached terminal state '%s', closing stream", orderID, order.Status)
				return nil
			}
		}
	}
}
