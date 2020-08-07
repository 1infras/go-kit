package redis

import (
	"context"
	"testing"
	"time"
)

func TestRedisUniversalClient(t *testing.T) {
	client, err := NewUniversalRedisClient(&Connection{
		Address: "localhost:6379",
	})

	if err != nil {
		t.Fatal(err)
	}

	client = Instrument(context.Background(), client)

	err = client.Set("foo", "bar", 1*time.Minute).Err()
	if err != nil {
		t.Fatal(err)
	}

	v, err := client.Get("foo").Result()
	if err != nil {
		t.Fatal(err)
	}

	if v != "bar" {
		t.Fatalf("expected is bar but actual is: %v", v)
	}
}
