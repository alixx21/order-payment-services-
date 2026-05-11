package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	jobKeyPrefix = "notification:job:"
	jobTTL       = 24 * time.Hour
)

type RedisJobIdempotency struct {
	client *redis.Client
}

func NewRedisJobIdempotency(client *redis.Client) *RedisJobIdempotency {
	return &RedisJobIdempotency{client: client}
}

func (r *RedisJobIdempotency) IsJobDone(ctx context.Context, eventID string) (bool, error) {
	val, err := r.client.Get(ctx, jobKeyPrefix+eventID).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // not found = not done
	}
	if err != nil {
		return false, fmt.Errorf("redis get: %w", err)
	}
	return val == "done", nil
}

func (r *RedisJobIdempotency) MarkJobDone(ctx context.Context, eventID string) error {
	return r.client.Set(ctx, jobKeyPrefix+eventID, "done", jobTTL).Err()
}
