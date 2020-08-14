package cache

import (
	"encoding/json"
	"testing"

	// "github.com/go-redis/redis"
	"github.com/1infras/go-kit/cache/redis"
)

func TestOneCache_GetInt64(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().SetRedis(r)

	cache.Set("hello", 1, 60)

	item, err := cache.Get("hello")
	if err != nil {
		t.Error(err)
		return
	}

	if item.Float64() != 1 {
		t.Error(err)
		return
	}
}

func TestOneCache_GetFloat(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().SetRedis(r)

	cache.Set("hello", 1.1, 60)

	val, err := cache.Get("hello")
	if err != nil {
		t.Error(err)
		return
	}

	if val.Float64() != 1.1 {
		t.Error(err)
		return
	}
}

func TestOneCache_GetString(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().SetRedis(r)

	cache.Set("hello_str", "Hello String", 60)

	val, err := cache.Get("hello_str")
	if err != nil {
		t.Error(err)
		return
	}

	if val.String() != "Hello String" {
		t.Error(err)
		return
	}
}

func TestOneCache_GetStruct(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	type fakeStruct struct {
		A int     `json:"a"`
		B string  `json:"b"`
		C float64 `json:"c"`
	}

	m := fakeStruct{1, "as", 2.0}

	cache := NewOneCacheStruct().SetRedis(r)

	cache.Set("hello_struct", m, 60)

	val, err := cache.Get("hello_struct")
	if err != nil {
		t.Error(err)
		return
	}

	dest := fakeStruct{}

	err = json.Unmarshal(val.Bytes(), &dest)
	if err != nil {
		t.Error(err)
		return
	}

	if dest.A != m.A {
		t.Error("Not equal")
		return
	}

	// t.Error(m)

	// t.Error(val.Value.([]byte))

}
