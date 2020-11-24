package cache

import (
	"time"

	"github.com/1infras/go-kit/lib/cache/lru"
	"github.com/1infras/go-kit/lib/cache/onecache"
	"github.com/1infras/go-kit/tracing"
	"github.com/1infras/go-kit/tracing/apmlru"
	"github.com/1infras/go-kit/tracing/hook"
)

func NewLRU(size int, expiration time.Duration) (lru.Client, error) {
	c, err := lru.NewWithExpiration(size, expiration)
	if err != nil {
		return nil, err
	}

	if tracing.Enabled {
		c.AddHook(apmlru.NewHook())
	}

	return c, err
}

func NewOneCache(opts ...onecache.ClientOptionFunc) (onecache.OneCache, error) {
	c, err := onecache.NewOneCache(opts...)
	if err != nil {
		return nil, err
	}

	if tracing.Enabled {
		c.AddHook(hook.NewHook())
	}

	return c, nil
}