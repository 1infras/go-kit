package cache

import (
	"context"
	"github.com/1infras/go-kit/src/cmd/cache/lru"
	"github.com/go-redis/redis"
	"go.elastic.co/apm"
	"reflect"
	"sync"
	"time"
)

const (
	//DefaultLRUSize
	DefaultLRUSize = 500
)

//MultiCache
//Caching with multiple layers include LRU and Redis
type MultiCache struct {
	LRU   *lru.Cache
	Redis redis.UniversalClient
	lock  sync.RWMutex
}

//NewMultiCache
//New a multi cache layers
func NewMultiCache(size int, expiration time.Duration, redisCache redis.UniversalClient) (*MultiCache, error) {
	if size <= 0 {
		size = DefaultLRUSize
	}

	l, err := lru.NewWithExpiration(size, expiration)
	if err != nil {
		return nil, err
	}

	return &MultiCache{
		LRU:   l,
		Redis: redisCache,
	}, nil
}

//Set cache to redis
func (c *MultiCache) redisSet(ctx context.Context, key string, values []byte, expiration time.Duration) error {
	span, _ := apm.StartSpan(ctx, "redis.set", "cache.multi_cache")
	defer span.End()
	return c.Redis.Set(key, values, expiration).Err()
}

//Set cache to LRU
func (c *MultiCache) lruSet(ctx context.Context, key string, values []byte, expiration time.Duration) bool {
	span, _ := apm.StartSpan(ctx, "lru.set", "cache.multi_cache")
	defer span.End()
	return c.LRU.Add(key, values, expiration)
}

//Get cache from redis
func (c *MultiCache) redisGet(ctx context.Context, key string) ([]byte, error) {
	span, _ := apm.StartSpan(ctx, "redis.get", "cache.multi_cache")
	defer span.End()
	return c.Redis.Get(key).Bytes()
}

//Get ttl cache from redis
func (c *MultiCache) redisGetTTL(ctx context.Context, key string) (time.Duration, error) {
	span, _ := apm.StartSpan(ctx, "redis.get_ttl", "cache.multi_cache")
	defer span.End()
	return c.Redis.TTL(key).Result()
}

//Get cache from lru
func (c *MultiCache) lruGet(ctx context.Context, key string) (interface{}, bool) {
	span, _ := apm.StartSpan(ctx, "lru.get", "cache.multi_cache")
	defer span.End()
	return c.LRU.Get(key)
}

//Set cache with value and expiration
func (c *MultiCache) Set(key string, values []byte, expiration time.Duration) (bool, error) {
	return c.SetCtx(nil, key, values, expiration)
}

//Get cache with multiple layers (LRU first, backed by Redis)
func (c *MultiCache) Get(key string) ([]byte, error) {
	//Get from LRU
	values, ok := c.LRU.Get(key)
	if ok {
		v := reflect.ValueOf(values)
		return v.Bytes(), nil
	}

	//Get from Redis
	v, err := c.Redis.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	//Get TTL from Redis
	ttl, err := c.Redis.TTL(key).Result()
	if err != nil {
		return nil, err
	}

	//Set back to LRU with TTL
	c.LRU.Add(key, v, ttl)

	return v, nil
}

//SetCtx cache with value and expiration, support transaction to tracing
func (c *MultiCache) SetCtx(ctx context.Context, key string, values []byte, expiration time.Duration) (bool, error) {
	//Set to redis first
	err := c.redisSet(ctx, key, values, expiration)
	if err != nil {
		return false, err
	}

	//Set to LRU
	c.lruSet(ctx, key, values, expiration)

	return true, nil
}

//Get cache with multiple layers (LRU first, backed by Redis), support transaction to tracing
func (c *MultiCache) GetCtx(ctx context.Context, key string) ([]byte, error) {
	//Get from LRU
	values, ok := c.lruGet(ctx, key)
	if ok {
		v := reflect.ValueOf(values)
		return v.Bytes(), nil
	}

	//Get from Redis
	v, err := c.redisGet(ctx, key)
	if err != nil {
		return nil, err
	}

	//Get TTL from Redis
	ttl, err := c.redisGetTTL(ctx, key)
	if err != nil {
		return nil, err
	}

	//Set back to LRU with TTL
	c.lruSet(ctx, key, v, ttl)

	return v, nil
}
