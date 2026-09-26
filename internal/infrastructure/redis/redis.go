package redis

import (
	"context"
	"ecommerce/config"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.RedisConfig, log *slog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Error("failed to connect to redis", "error", err)
		panic(fmt.Sprintf("failed to connect to redis: %s", err))
	}
	log.Info("redis connection established")
	return client
}
