package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])

if current == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
`)
func (r *RedisRateLimiter) Allow(ctx context.Context,key string,limit int,windowSeconds int)(bool,error){
	redisKey:=fmt.Sprintf("rate_limit:%s",key)
	result,err:=rateLimitScript.Run(ctx,r.client,[]string{redisKey},windowSeconds).Int64()
	if err != nil{
		return false,err
	}
	return result <= int64(limit),nil
}
