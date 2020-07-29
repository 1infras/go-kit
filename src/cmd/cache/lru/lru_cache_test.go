package lru

import "testing"

func TestCache(t *testing.T) {
	var cache LruCache = New(2)
	//1 will be added
	if cache.Add(1, 1) {
		t.Error("should not have an eviction")
	}
	//2 will be added
	cache.Add(2, 2)
	//1 was evicted, and 3 will be added
	if !cache.Add(3, 3) {
		t.Error("should have an eviction")
	}
	//However, the key is still keep, but it's not in the list of recent-ness cache
	if cache.Contains(1) {
		t.Error("key must be exist")
	}

	//Try get value cache
	v, ok := cache.Get(2)
	if !ok {
		t.Error("cache must be exist")
	}
	if v != 2 {
		t.Error("cache value must be equal 2")
	}

	//Try remove
	cache.Remove(2)
	_, ok = cache.Get(2)
	if ok {
		t.Error("cache must not be exist")
	}

	//Get length
	if cache.Len() != 1 {
		t.Error("size must equal 1")
	}

	//Purge cache
	cache.Purge()
	if cache.Len() != 0 {
		t.Error("cache must be purged")
	}

	//Try add cache and peek
	cache.Add(1, 1)
	cache.Add(2, 2)
	if v, ok := cache.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	//Try evict the oldest cache
	cache.Add(3, 3)
	if cache.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}

	//Print key after evicted
	keys := cache.GetKeys()
	t.Log("Recently cache keys")
	for _, k := range keys {
		t.Log(k)
	}
}
