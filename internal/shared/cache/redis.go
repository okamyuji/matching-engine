package cache

import (
	"context"
	"time"
)

// Cache キャッシュインターフェース
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Redis実装は将来追加可能
// go get github.com/redis/go-redis/v9 を実行後、以下のコードを有効化

/*
import "github.com/redis/go-redis/v9"

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, db int) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &redisCache{client: client}
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
*/
