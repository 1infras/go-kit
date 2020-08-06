package lru

import lru "github.com/hashicorp/golang-lru"

//Cache -
type Cache struct {
	C *lru.Cache
}

//New - New creates an LRU of the given size (size = maximum elements).
func New(size int) *Cache {
	c, _ := lru.New(size)
	return &Cache{C: c}
}

//Add - Add adds a value to the cache. Returns true if an eviction occurred.
func (c *Cache) Add(key, value interface{}) bool {
	return c.C.Add(key, value)
}

//Get - Get looks up a key's value from the cache.
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	return c.C.Get(key)
}

//Contains - Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *Cache) Contains(key interface{}) bool {
	return c.C.Contains(key)
}

//Peek - Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *Cache) Peek(key interface{}) (interface{}, bool) {
	return c.C.Peek(key)
}

//Remove - Remove removes the provided key from the cache.
func (c *Cache) Remove(key interface{}) bool {
	return c.C.Remove(key)
}

//GetKeys - Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache) GetKeys() []interface{} {
	return c.C.Keys()
}

//Len - Len returns the number of items in the cache.
func (c *Cache) Len() int {
	return c.C.Len()
}

//Purge - Purge is used to completely clear the cache.
func (c *Cache) Purge() {
	c.C.Purge()
}
