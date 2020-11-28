package lru

import (
	"context"
	"sync"
	"time"

	"github.com/1infras/go-kit/lib/cache/lru/core"
	"github.com/1infras/go-kit/lib/hook/common"
)

// Client is a thread-safe fixed size LRU cache.
type Client interface {
	Purge(ctx context.Context)
	Add(ctx context.Context, key, value interface{}, expiration time.Duration) (evicted bool)
	Get(ctx context.Context, key interface{}) (value interface{}, ok bool)
	Contains(ctx context.Context, key interface{}) bool
	Peek(ctx context.Context, key interface{}) (value interface{}, ok bool)
	ContainsOrAdd(ctx context.Context, key, value interface{}, expiration time.Duration) (ok, evicted bool)
	PeekOrAdd(ctx context.Context, key, value interface{}, expiration time.Duration) (previous interface{}, ok, evicted bool)
	Remove(ctx context.Context, key interface{}) (present bool)
	Resize(ctx context.Context, size int) (evicted int)
	RemoveOldest(ctx context.Context) (key, value interface{}, ok bool)
	GetOldest(ctx context.Context) (key, value interface{}, ok bool)
	Keys(ctx context.Context) []interface{}
	Len(ctx context.Context) int
	AddHook(hook common.HookProcess)
}

type lru struct {
	lru  core.LRUCache
	lock sync.RWMutex
	hook *common.Hook
}

// New creates an LRU of the given size.
func New(size int) (Client, error) {
	return NewWithEvict(size, nil)
}

// NewWithExpiration creates an LRU of the given size with a default expiration time for every item
func NewWithExpiration(size int, expiration time.Duration) (Client, error) {
	return NewWithEvictExpiration(size, expiration, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(size int, onEvicted func(key interface{}, value interface{})) (Client, error) {
	return NewWithEvictExpiration(size, 0, onEvicted)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvictExpiration(size int, expiration time.Duration, onEvicted func(key interface{}, value interface{})) (Client, error) {
	client, err := core.NewLRUWithExpire(size, expiration, onEvicted)
	if err != nil {
		return nil, err
	}
	c := &lru{
		lru:  client,
		hook: &common.Hook{},
	}

	return c, nil
}

// AddHook is used wrap Processing before and after for a Process
func (_this *lru) AddHook(hook common.HookProcess) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.AddHook(hook)
}

// Purge is used to completely clear the cache.
func (_this *lru) Purge(ctx context.Context) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		_this.lru.Purge()
	}, "purge")
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (_this *lru) Add(ctx context.Context, key, value interface{}, expiration time.Duration) (evicted bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		evicted = _this.lru.Add(key, value, expiration)
	}, "add")
	return
}

// Get looks up a key's value from the cache.
func (_this *lru) Get(ctx context.Context, key interface{}) (value interface{}, ok bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		value, ok = _this.lru.Get(key)
	}, "get")
	return
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (_this *lru) Contains(ctx context.Context, key interface{}) (existed bool) {
	_this.lock.RLock()
	defer _this.lock.RUnlock()

	_this.hook.Process(ctx, func() {
		existed = _this.lru.Contains(key)
	}, "contains")
	return
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (_this *lru) Peek(ctx context.Context, key interface{}) (value interface{}, ok bool) {
	_this.lock.RLock()
	defer _this.lock.RUnlock()

	_this.hook.Process(ctx, func() {
		value, ok = _this.lru.Peek(key)
	}, "peek")
	return
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (_this *lru) ContainsOrAdd(ctx context.Context, key, value interface{}, expiration time.Duration) (ok, evicted bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		ok = _this.lru.Contains(key)
	}, "contains")

	if ok {
		return
	}

	_this.hook.Process(ctx, func() {
		evicted = _this.lru.Add(key, value, expiration)
	}, "add")
	return
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (_this *lru) PeekOrAdd(ctx context.Context, key, value interface{}, expiration time.Duration) (previous interface{}, ok, evicted bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		previous, ok = _this.lru.Peek(key)
	}, "peek")

	if ok {
		return
	}

	_this.hook.Process(ctx, func() {
		evicted = _this.lru.Add(key, value, expiration)
	}, "add")
	return
}

// Remove removes the provided key from the cache.
func (_this *lru) Remove(ctx context.Context, key interface{}) (present bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		present = _this.lru.Remove(key)
	}, "remove")
	return
}

// Resize changes the cache size.
func (_this *lru) Resize(ctx context.Context, size int) (evicted int) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		evicted = _this.lru.Resize(size)
	}, "resize")
	return
}

// RemoveOldest removes the oldest item from the cache.
func (_this *lru) RemoveOldest(ctx context.Context) (key, value interface{}, ok bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		key, value, ok = _this.lru.RemoveOldest()
	}, "remove_oldest")
	return
}

// GetOldest returns the oldest entry
func (_this *lru) GetOldest(ctx context.Context) (key, value interface{}, ok bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		key, value, ok = _this.lru.GetOldest()
	}, "get_oldest")
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (_this *lru) Keys(ctx context.Context) (keys []interface{}) {
	_this.lock.RLock()
	defer _this.lock.RUnlock()

	_this.hook.Process(ctx, func() {
		keys = _this.lru.Keys()
	}, "keys")
	return
}

// Len returns the number of items in the cache.
func (_this *lru) Len(ctx context.Context) (length int) {
	_this.lock.RLock()
	defer _this.lock.RUnlock()

	_this.hook.Process(ctx, func() {
		length = _this.lru.Len()
	}, "len")
	return
}
