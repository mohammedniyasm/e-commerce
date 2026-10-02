package interfaces

import "context"

type RateLimiter interface {
	Allow(ctx context.Context,key string,limit int,windowSeconds int)(bool,error)
}