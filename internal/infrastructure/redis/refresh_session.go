package redis

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshSessionStore struct {
	client        *redis.Client
	refreshExpiry time.Duration
}

func NewRefreshSessionStore(client *redis.Client, refreshExpiry time.Duration) *RefreshSessionStore {
	return &RefreshSessionStore{
		client:        client,
		refreshExpiry: refreshExpiry,
	}
}
func (s *RefreshSessionStore) Save(ctx context.Context, jti string, userID uint) error {
	key := "refresh_session:" + jti
	return s.client.Set(ctx, key, strconv.FormatUint(uint64(userID), 10), s.refreshExpiry).Err()
}
func (s *RefreshSessionStore) Get(ctx context.Context, jti string) (uint, error) {
	key := "refresh_session:" + jti

	value, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, domainerrors.ErrRefreshSessionNotFound
		}
		return 0, err
	}

	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}
func (s *RefreshSessionStore) Delete(ctx context.Context, jti string) error {
	key := "refresh_session:" + jti

	return s.client.Del(ctx, key).Err()
}
var consumeRefreshSessionScript = redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if not value then
	return ""
end
redis.call("DEL", KEYS[1])
return value
`)
func (r *RefreshSessionStore) Consume(ctx context.Context,jti string) (uint, bool, error) {
	key := "refresh_session:" + jti

	value, err := consumeRefreshSessionScript.Run(ctx,r.client,[]string{key}).Text()
	if err != nil {
		return 0, false, err
	}
	if value == "" {
		return 0, false, nil
	}
	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return uint(userID), true, nil
}
