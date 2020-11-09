package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/1infras/go-kit/driver/redis"
)

func TestMultiCache(t *testing.T) {
	r, err := redis.NewDefaultRedisUniversalClient()
	assert.Nil(t, err)

	c, err := NewMultiCache(100, 5*time.Second, r)
	assert.Nil(t, err)

	_, err = c.Set("foo", []byte("bar"), 5*time.Second)
	assert.Nil(t, err)

	v, err := c.Get("foo")
	assert.Nil(t, err)

	if p := string(v); p != "bar" {
		t.Fatalf("Expected is bar but actual is %v", p)
	}

	time.Sleep(2 * time.Second)
	v, err = c.Get("foo")
	assert.Nil(t, err)

	assert.Equal(t, "bar", string(v))

	time.Sleep(5 * time.Second)

	v, err = c.Get("foo")
	assert.NotNil(t, err)
}
