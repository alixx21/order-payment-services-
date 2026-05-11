package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	rediscache "notification-service/internal/cache/redis"
	"notification-service/internal/consumer/rabbitmq"
	"notification-service/internal/provider"
	"notification-service/internal/provider/simulated"
	"notification-service/internal/repository/postgres"
	"notification-service/internal/usecase"
	"notification-service/internal/worker"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, reading from environment")
	}

	dsn := getEnv("DATABASE_URL", "postgres://postgres:123123@localhost:5433/notifications_db?sslmode=disable")
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	providerMode := getEnv("PROVIDER_MODE", "SIMULATED")
	maxRetries := 4

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to notifications database")

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()
	log.Println("Connected to Redis")

	var emailSender provider.EmailSender
	switch providerMode {
	case "SIMULATED":
		emailSender = simulated.NewSimulatedEmailSender()
		log.Println("[Provider] Using SIMULATED email sender")
	default:
		log.Fatalf("unknown PROVIDER_MODE: %s (supported: SIMULATED)", providerMode)
	}

	idempotencyRepo := postgres.NewIdempotencyRepository(db)
	jobIdempotency := rediscache.NewRedisJobIdempotency(redisClient)
	notificationWorker := worker.NewWorker(emailSender, maxRetries)
	notificationUC := usecase.New(idempotencyRepo, jobIdempotency, notificationWorker)

	consumer, err := rabbitmq.NewConsumer(rabbitmqURL, notificationUC)
	if err != nil {
		log.Fatalf("create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down notification service...")
	cancel()
	log.Println("Notification service stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
