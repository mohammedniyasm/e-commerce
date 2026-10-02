package interfaces

import (
	"context"
	"time"
)

type AccessTokenBlacklist interface {
	Blacklist(ctx context.Context, jti string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}
