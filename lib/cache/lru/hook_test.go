package lru

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

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

func TestHook_AddHook(t *testing.T) {
	logger.InitLogger(logger.InfoLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := New(100)
	assert.Nil(t, err)
	l.AddHook(&testHook{})

	l.Add(ctx, "hello", "1", 0)
	v, ok := l.Get(ctx, "hello")
	assert.Equal(t, true, ok)
	assert.Equal(t, "1", v)
}
