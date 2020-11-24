package consume

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/1infras/go-kit/driver/kafka"
	"github.com/1infras/go-kit/logger"
)

type testHook struct{}

func (*testHook) BeforeProcess(ctx context.Context, cmdName string) context.Context {
	ctx = context.WithValue(ctx, "start_time", time.Now().Unix())
	return ctx
}

func (*testHook) AfterProcess(ctx context.Context, cmdName string) {
	start := time.Unix(reflect.ValueOf(ctx.Value("start_time")).Int(), 0)
	logger.Info("consumed with a time", zap.String("cmd", cmdName), zap.String("duration", time.Since(start).String()))
}

func TestConsume(t *testing.T) {
	logger.InitLogger(logger.DebugLevel)
	k, err := kafka.NewKafka(&kafka.Config{
		Brokers: []string{"localhost:9092"},
	})

	assert.Nil(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	c, err := CreateConsumer(k,
		SetTopic("my-topic"),
		SetContext(ctx),
		SetBalanceStrategyMode(Sticky),
		SetGroup("my-group"),
		SetInitialOffsetMode(Oldest),
		SetClose(func() {
			logger.Info("closing the consumer")
		}))

	assert.Nil(t, err)
	c.AddHook(&testHook{})
	c.SetConsumeHandler(func(message []byte) ConsumeStatus {
		var msg string
		err := json.Unmarshal(message, &msg)
		if err != nil {
			logger.Error("malformed message, couldn't unmarshall", zap.String("message", string(message)))
			return Dispeard
		}
		logger.Info("consumed success", zap.String("result", msg))

		t.Logf("report: %s", c.Report())
		return Consumed
	})

	c.Run()
}
