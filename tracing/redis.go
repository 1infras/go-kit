package tracing

import (
	"github.com/go-redis/redis/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
)

func WrapGoRedisClient(r *redis.Client) {
	r.AddHook(apmgoredis.NewHook())
}

func WrapGoRedisClusterClient(r *redis.ClusterClient) {
	r.AddHook(apmgoredis.NewHook())
}

func WrapGoRedisRing(r *redis.Ring) {
	r.AddHook(apmgoredis.NewHook())
}

func WrapGoRedisUniversalClient(r redis.UniversalClient) {
	r.AddHook(apmgoredis.NewHook())
}
