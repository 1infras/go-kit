package tracing

import (
	"github.com/go-redis/redis/v8"
	"go.elastic.co/apm/module/apmgoredisv8"
)

func WrapGoRedisClient(r *redis.Client) {
	r.AddHook(apmgoredisv8.NewHook())
}

func WrapGoRedisClusterClient(r *redis.ClusterClient) {
	r.AddHook(apmgoredisv8.NewHook())
}

func WrapGoRedisUniversalClient(r redis.UniversalClient) {
	r.AddHook(apmgoredisv8.NewHook())
}