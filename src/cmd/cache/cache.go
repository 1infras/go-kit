package cache

import (
	"github.com/1infras/go-kit/src/cmd/cache/lru"
	"github.com/go-redis/redis"
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

//Set
//Set cache with value and expiration
func (c *MultiCache) Set(key string, values []byte, expiration time.Duration) (bool, error) {
	//Set to redis first
	err := c.Redis.Set(key, values, expiration).Err()
	if err != nil {
		return false, err
	}

	//Set to LRU
	c.LRU.Add(key, values, expiration)

	return true, nil
}

//Get
//Get cache with multiple layers (LRU first, backed by Redis)
func (c *MultiCache) Get(key string) ([]byte, error) {
	//Get from LRU
	values, ok := c.LRU.Get(key)
	if ok {
		v := reflect.ValueOf(values)
		return v.Bytes(), nil
	}

	v, err := c.Redis.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	ttl, err := c.Redis.TTL(key).Result()
	if err != nil {
		return nil, err
	}

	c.LRU.Add(key, v, ttl)

	return v, nil
}
