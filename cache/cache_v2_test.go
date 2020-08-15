package cache

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/1infras/go-kit/cache/redis"
	// "github.com/go-redis/redis"
)

func TestOneCache_GetInt64(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().WithRedis(r, "hello")

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

	cache := NewOneCacheStruct().WithRedis(r, "hello")

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

	cache := NewOneCacheStruct().WithRedis(r, "hello")

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

	m := &fakeStruct{1, "as", 2.0}

	cache := NewOneCacheStruct().WithRedis(r, "hello")

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

}

func TestOneCacheStruct_Flush(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().WithRedis(r, "hello")
	cache.Flush()
}

func TestOneCacheStruct_Report(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	if err != nil {
		t.Fatal(err)
	}

	cache := NewOneCacheStruct().WithRedis(r, "hello")
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("200_%v", i), i, 60)
	}

	cache.Get("200_1")
	cache.Get("100")

	fmt.Println(cache.Report())

}
