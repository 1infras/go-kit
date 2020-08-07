package cache

import (
	"context"
	"time"
)

type Cache interface {
	//Set cache with optional expiration time, set 0 if the cache would be infinity
	Set(key string, values []byte, expiration time.Duration) (bool, error)
	//Get cache by key
	Get(key string) ([]byte, error)
	//SetCtx cache with optional expiration time, set 0 if the cache would be infinity, support tracing
	SetCtx(ctx context.Context, key string, values []byte, expiration time.Duration) (bool, error)
	//Get cache by key, support tracing
	GetCtx(ctx context.Context, key string) ([]byte, error)
}
