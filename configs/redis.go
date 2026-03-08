package configs

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

var RedisClient = redis.NewClient(&redis.Options{
	Addr:     getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
	Password: getEnv("REDIS_PASSWORD", ""),
	DB:       0,
})

func PingRedis() error {
	return RedisClient.Ping(ctx).Err()
}
