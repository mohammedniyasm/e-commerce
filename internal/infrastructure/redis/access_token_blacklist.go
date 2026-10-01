package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)
const accessTokenBlacklistPrefix="blacklisted_access:"
type AccessTokenBlacklistStore struct {
	client *redis.Client
}

func NewAccessTokenBlacklistStore(client *redis.Client) *AccessTokenBlacklistStore {
	return &AccessTokenBlacklistStore{
		client: client,
	}
}
func (s AccessTokenBlacklistStore) Blacklist(ctx context.Context, jti string, expiration time.Duration)error{
	key:=accessTokenBlacklistPrefix+jti
	return s.client.Set(ctx,key,jti,expiration).Err()
}
func (s *AccessTokenBlacklistStore) IsBlacklisted(ctx context.Context,jti string)(bool,error){
	key:=accessTokenBlacklistPrefix+jti
	exists,err:=s.client.Exists(ctx,key).Result()
	if err != nil{
		return false,err
	}
	return exists>0,nil
}
