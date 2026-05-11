package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		ctx := c.Request.Context()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(limit)-count)))

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("rate limit exceeded: max %d requests per %s", limit, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
