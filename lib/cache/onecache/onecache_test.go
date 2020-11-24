package onecache

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/1infras/go-kit/driver/redis"
	"github.com/1infras/go-kit/logger"
)

type testHook struct{}

func (*testHook) BeforeProcess(ctx context.Context, cmdName string) context.Context {
	ctx = context.WithValue(ctx, "start_time", time.Now().Unix())
	return ctx
}

func (*testHook) AfterProcess(ctx context.Context, cmdName string) {
	start := time.Unix(reflect.ValueOf(ctx.Value("start_time")).Int(), 0)
	logger.Info("Total time complete transaction", zap.String("cmd", cmdName), zap.String("duration", time.Since(start).String()))
}

func TestOneCache(t *testing.T) {
	logger.InitLogger(logger.InfoLevel)
	ctx := context.Background()

	redisCache, err := redis.NewDefaultRedisUniversalClient()
	assert.Nil(t, err)

	oneCache, err := NewOneCache(
		SetContext(ctx),
		SetExpiration(3*time.Second),
		SetMaxItems(10),
		SetRemoteCacheNamespace("redis_cache"),
		SetRemoteCache(redisCache))
	assert.Nil(t, err)

	oneCache.AddHook(&testHook{})
	err = oneCache.Set(ctx, "hello", 1, 3)
	assert.Nil(t, err)

	v, err := oneCache.Get(ctx, "hello")
	assert.Nil(t, err)

	vk, err := v.Int()
	assert.Nil(t, err)
	assert.Equal(t, 1, vk)

	existed := oneCache.Contains(ctx, "hello")
	assert.Equal(t, true, existed)

	oneCache.Delete(ctx, "hello")

	_, err = oneCache.Get(ctx, "hello")
	assert.NotNil(t, err)
	assert.Equal(t, Nil.Error(), err.Error())

	report := oneCache.Report(ctx)
	t.Logf("%s", report)

	err = oneCache.Flush(ctx)
	assert.Nil(t, err)
}