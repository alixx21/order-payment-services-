package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	orderpb "order-service/internal/orderpb"
	pb "order-service/internal/pb"
	"order-service/internal/repository/postgres"
	grpcserver "order-service/internal/transport/grpc"
	transport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	dsn := getEnv("DATABASE_URL", "postgres://postgres:123123@localhost:5433/orders_db?sslmode=disable")
	httpPort := getEnv("PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "9090")
	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:9091")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to orders database")

	conn, err := grpc.NewClient(
		paymentGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect to payment grpc: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to Payment gRPC at %s", paymentGRPCAddr)

	orderRepo := postgres.NewOrderRepository(db)
	paymentClient := transport.NewPaymentGRPCClient(pb.NewPaymentServiceClient(conn))
	orderUC := usecase.New(orderRepo, paymentClient)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen on grpc port: %v", err)
		}
		grpcSrv := grpc.NewServer()
		orderpb.RegisterOrderServiceServer(grpcSrv, grpcserver.NewOrderGRPCServer(orderUC))
		log.Printf("Order gRPC streaming server listening on :%s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()
	handler := transport.NewOrderHandler(orderUC)
	r := gin.Default()
	handler.RegisterRoutes(r)
	log.Printf("Order Service HTTP listening on :%s", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
