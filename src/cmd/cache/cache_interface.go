package cache

import "time"

type Cache interface {
	//Set cache with optional expiration time, set 0 if the cache would be infinity
	Set(key string, values []byte, expiration time.Duration) (bool, error)
	//Get cache by key
	Get(key string) ([]byte, error)
}
