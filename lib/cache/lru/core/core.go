package core

import (
	"container/list"
	"errors"
	"time"
)

// LRUCache is the interface for simple LRU cache.
type LRUCache interface {
	// Adds a value to the cache, returns true if an eviction occurred and
	// updates the "recently used"-ness of the key.
	Add(key, value interface{}, expiration time.Duration) bool

	// Returns key's value from the cache and
	// updates the "recently used"-ness of the key. #value, isFound
	Get(key interface{}) (value interface{}, ok bool)

	// Checks if a key exists in cache without updating the recent-ness.
	Contains(key interface{}) (ok bool)

	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(key interface{}) (value interface{}, ok bool)

	// Removes a key from the cache.
	Remove(key interface{}) bool

	// Removes the oldest entry from cache.
	RemoveOldest() (interface{}, interface{}, bool)

	// Returns the oldest entry from the cache. #key, value, isFound
	GetOldest() (interface{}, interface{}, bool)

	// Returns a slice of the keys in the cache, from oldest to newest.
	Keys() []interface{}

	// Returns the number of items in the cache.
	Len() int

	// Clears all cache entries.
	Purge()

	// Resizes cache, returning number evicted
	Resize(int) int
}

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   EvictCallback
	expire    time.Duration
}

// entry is used to hold a value in the evictList
type entry struct {
	key    interface{}
	value  interface{}
	expire *time.Time
}

// New constructs an LRU of the given size
func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	return NewLRUWithExpire(size, 0, onEvict)
}

// NewLRU constructs an LRU of the given size with default expire
func NewLRUWithExpire(size int, expire time.Duration, onEvict EvictCallback) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
		expire:    expire,
	}
	return c, nil
}

// IsExpire is used to check expiration of cache
func (_this *entry) IsExpire() bool {
	if _this.expire == nil {
		return false
	}
	return time.Now().After(*_this.expire)
}

// Purge is used to completely clear the cache.
func (_this *LRU) Purge() {
	for k, v := range _this.items {
		if _this.onEvict != nil {
			_this.onEvict(k, v.Value.(*entry).value)
		}
		delete(_this.items, k)
	}
	_this.evictList.Init()
}

// Add add a value to the cache with optional expiration. Returns true if an eviction occurred.
func (_this *LRU) Add(key, value interface{}, expiration time.Duration) (evicted bool) {
	var ex *time.Time = nil

	if expiration > 0 {
		expire := time.Now().Add(expiration)
		ex = &expire
	} else if _this.expire > 0 {
		expire := time.Now().Add(_this.expire)
		ex = &expire
	}

	// Check for existing item
	if ent, ok := _this.items[key]; ok {
		_this.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		ent.Value.(*entry).expire = ex
		return false
	}

	// Add new item
	ent := &entry{key: key, value: value, expire: ex}
	entry := _this.evictList.PushFront(ent)
	_this.items[key] = entry

	evict := _this.evictList.Len() > _this.size
	// Verify size not exceeded
	if evict {
		_this.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (_this *LRU) Get(key interface{}) (value interface{}, ok bool) {
	if ent, ok := _this.items[key]; ok {
		if ent.Value.(*entry).IsExpire() {
			return nil, false
		}
		_this.evictList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (_this *LRU) Contains(key interface{}) (ok bool) {
	if ent, ok := _this.items[key]; ok {
		if ent.Value.(*entry).IsExpire() {
			return false
		}
		return ok
	}
	return
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (_this *LRU) Peek(key interface{}) (value interface{}, ok bool) {
	var ent *list.Element
	if ent, ok = _this.items[key]; ok {
		if ent.Value.(*entry).IsExpire() {
			return nil, false
		}
		return ent.Value.(*entry).value, true
	}
	return nil, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (_this *LRU) Remove(key interface{}) (present bool) {
	if ent, ok := _this.items[key]; ok {
		_this.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (_this *LRU) RemoveOldest() (key, value interface{}, ok bool) {
	ent := _this.evictList.Back()
	if ent != nil {
		_this.removeElement(ent)
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// GetOldest returns the oldest entry
func (_this *LRU) GetOldest() (key, value interface{}, ok bool) {
	ent := _this.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (_this *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(_this.items))
	i := 0
	for ent := _this.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (_this *LRU) Len() int {
	return _this.evictList.Len()
}

// Resize changes the cache size.
func (_this *LRU) Resize(size int) (evicted int) {
	diff := _this.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		_this.removeOldest()
	}
	_this.size = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (_this *LRU) removeOldest() {
	ent := _this.evictList.Back()
	if ent != nil {
		_this.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (_this *LRU) removeElement(e *list.Element) {
	_this.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(_this.items, kv.key)
	if _this.onEvict != nil {
		_this.onEvict(kv.key, kv.value)
	}
}
