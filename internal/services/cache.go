package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheService struct {
	Client *redis.Client
}

func NewCacheService(addr string) *CacheService {
	return &CacheService{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		}),
	}
}

func (c *CacheService) Get(key string) ([]byte, error) {
	ctx := context.Background()
	return c.Client.Get(ctx, key).Bytes()
}

func (c *CacheService) Set(key string, value []byte, ttl time.Duration) error {
	ctx := context.Background()
	return c.Client.Set(ctx, key, value, ttl).Err()
}
