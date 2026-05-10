# Assignment 3 — Event-Driven Architecture (RabbitMQ)

## Architecture

```
User (Postman)
     │
     │ HTTP REST
     ▼
Order Service :8080
     │
     │ gRPC (unchanged from Assignment 2)
     ▼
Payment Service :8081/:9091
     │
     │ Publish event to RabbitMQ (after successful DB commit)
     ▼
RabbitMQ :5672 (queue: payment.completed)
     │
     │ Consume event (manual ACK)
     ▼
Notification Service
     │
     ▼
Console log: [Notification] Sent email to user@example.com for Order #123. Amount: $50.00
```

## What's New in Assignment 3

| Component | Change |
|---|---|
| Payment Service | Publishes `PaymentEvent` to RabbitMQ after successful payment |
| Notification Service | New microservice — consumes events from RabbitMQ |
| RabbitMQ | Message broker — queue `payment.completed`, durable, persistent messages |
| Docker Compose | All 5 services run with `docker-compose up --build` |

## Event Payload

```json
{
  "event_id": "uuid",
  "order_id": "string",
  "amount": 5000,
  "customer_email": "user@example.com",
  "status": "COMPLETED"
}
```

## How to Run

```bash
# From project root
docker-compose up --build

# Services:
# Order Service   → http://localhost:8080
# Payment Service → http://localhost:8081
# RabbitMQ UI     → http://localhost:15672 (guest/guest)
```

## ACK Logic

Manual ACK is used — message is only acknowledged after:
1. JSON parsed successfully
2. Idempotency check passes (event not already processed)
3. Event marked in DB as processed
4. Email log printed successfully

If any step fails → `NACK` with `requeue=false` (drops bad message, avoids infinite loop).

## Idempotency Strategy

PostgreSQL table `processed_events`:
```sql
CREATE TABLE processed_events (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Before processing any event:
1. Check if `event_id` exists in table
2. If yes → skip, ACK safely (duplicate)
3. If no → insert `event_id` → process → ACK

Using PostgreSQL (not in-memory) means duplicates detected even after service restart.

## How to Test

```bash
# 1. Start all services
docker-compose up --build

# 2. Create a successful order (amount <= 100000)
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"c1","item_name":"Book","amount":5000,"customer_email":"user@example.com"}'

# 3. Check Notification Service logs
docker logs notification-service
# Expected: [Notification] Sent email to user@example.com for Order #xxx. Amount: $50.00

# 4. Stop Notification Service
docker-compose stop notification-service

# 5. Create another payment (message stays in RabbitMQ queue)
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"c1","item_name":"Laptop","amount":75000,"customer_email":"user@example.com"}'

# 6. Start Notification Service again
docker-compose start notification-service

# 7. Check logs — message was NOT lost (RabbitMQ held it)
docker logs notification-service
# Message delivered and processed after restart
```

## Port Map

| Service | Protocol | Port |
|---|---|---|
| Order Service | HTTP REST | 8080 |
| Order Service | gRPC streaming | 9090 |
| Payment Service | HTTP REST | 8081 |
| Payment Service | gRPC | 9091 |
| RabbitMQ | AMQP | 5672 |
| RabbitMQ | Management UI | 15672 |
| PostgreSQL | SQL | 5433 |
