FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY ap2-generated ./ap2-generated
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/order-service

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8080
CMD ["./order-service"]