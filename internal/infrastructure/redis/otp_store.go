package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPStore struct {
	client *redis.Client
}

func NewOTPStore(client *redis.Client) *OTPStore {
	return &OTPStore{
		client: client,
	}
}
func (s *OTPStore) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}
func (s *OTPStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}
func (s *OTPStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
