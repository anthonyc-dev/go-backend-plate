package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rest-api/configs"
	"rest-api/utils"

	"github.com/gin-gonic/gin"
)

func RateLimiterMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		key := fmt.Sprintf("rate_limit:%s", clientIP)
		infoKey := fmt.Sprintf("rate_limit_info:%s", clientIP)

		ctx := context.Background()

		count, err := configs.RedisClient.Get(ctx, key).Int()
		if err != nil && err.Error() != "redis: nil" {
			utils.LogError(c, http.StatusInternalServerError, "RATE_LIMIT_ERROR", "Rate limiter error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			c.Abort()
			return
		}

		if count >= maxRequests {
			ttl, _ := configs.RedisClient.TTL(ctx, key).Result()
			utils.LogError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": ttl.Seconds(),
			})
			c.Abort()
			return
		}

		configs.RedisClient.Incr(ctx, key)
		if count == 0 {
			configs.RedisClient.Expire(ctx, key, window)

			info := map[string]interface{}{
				"ip":         clientIP,
				"user_agent": userAgent,
				"count":      1,
				"first_seen": time.Now().Unix(),
				"last_seen":  time.Now().Unix(),
			}
			for k, v := range info {
				configs.RedisClient.HSet(ctx, infoKey, k, fmt.Sprintf("%v", v))
			}
			configs.RedisClient.Expire(ctx, infoKey, window)
		} else {
			configs.RedisClient.HIncrBy(ctx, infoKey, "count", 1)
			configs.RedisClient.HSet(ctx, infoKey, "last_seen", fmt.Sprintf("%v", time.Now().Unix()))
		}

		c.Next()
	}
}
