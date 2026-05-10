package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/consumer/rabbitmq"
	"notification-service/internal/repository/postgres"
	"notification-service/internal/usecase"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading from environment")
	}

	dsn         := getEnv("DATABASE_URL", "postgres://postgres:123123@localhost:5433/notifications_db?sslmode=disable")
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// --- Database ---
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to notifications database")

	// ===========================================================
	// Composition Root
	// ===========================================================
	idempotencyRepo := postgres.NewIdempotencyRepository(db)
	notificationUC  := usecase.New(idempotencyRepo)

	consumer, err := rabbitmq.NewConsumer(rabbitmqURL, notificationUC)
	if err != nil {
		log.Fatalf("create consumer: %v", err)
	}
	defer consumer.Close()

	// --- Graceful Shutdown ---
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming in background goroutine
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	// Block until OS signal received
	<-quit
	log.Println("Shutting down notification service...")
	cancel() // stop the consumer goroutine
	log.Println("Notification service stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
