package lru

//A wrapper LruCache by Hashicorp
type LruCache interface {
	// Adds a value to the cache, returns true if an eviction occurred and
	// updates the "recently used"-ness of the key.
	Add(key, value interface{}) bool
	// Returns key's value from the cache and
	// updates the "recently used"-ness of the key. #value, isFound
	Get(key interface{}) (interface{}, bool)
	// Checks if a key exists in cache without updating the recent-ness.
	Contains(key interface{}) bool
	// Returns key's value without updating the "recently used"-ness of the key.
	Peek(key interface{}) (interface{}, bool)
	// Removes a key from the cache.
	Remove(key interface{}) bool
	// Returns a slice of the keys in the cache, from oldest to newest.
	GetKeys() []interface{}
	// Returns the number of items in the cache.
	Len() int
	// Clears all cache entries.
	Purge()
}
