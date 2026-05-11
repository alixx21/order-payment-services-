package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"order-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "order:"

type OrderRedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewOrderRedisCache(client *redis.Client, ttl time.Duration) *OrderRedisCache {
	return &OrderRedisCache{client: client, ttl: ttl}
}

func (c *OrderRedisCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	val, err := c.client.Get(ctx, keyPrefix+id).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("cache miss") // caller will query DB
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var order domain.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, fmt.Errorf("unmarshal order: %w", err)
	}
	return &order, nil
}

func (c *OrderRedisCache) Set(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}
	return c.client.Set(ctx, keyPrefix+order.ID, data, c.ttl).Err()
}

func (c *OrderRedisCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, keyPrefix+id).Err()
}
