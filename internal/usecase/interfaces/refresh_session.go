package interfaces

import "context"

type RefreshSessionStore interface {
	Save(ctx context.Context, jti string, userID uint) error
	Get(ctx context.Context, jti string) (uint, error)
	Delete(ctx context.Context, jti string) error
	Consume(ctx context.Context, jti string) (uint, bool, error)
}
