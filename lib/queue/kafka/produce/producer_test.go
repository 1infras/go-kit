package produce

import (
	"context"
	"fmt"
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
	// nolint:staticcheck
	ctx = context.WithValue(ctx, "start_time", time.Now().Unix())
	return ctx
}

func (*testHook) AfterProcess(ctx context.Context, cmdName string) {
	start := time.Unix(reflect.ValueOf(ctx.Value("start_time")).Int(), 0)
	logger.Info("Total time complete transaction", zap.String("cmd", cmdName), zap.String("duration", time.Since(start).String()))
}

func TestSyncProduce(t *testing.T) {
	logger.InitLogger(logger.DebugLevel)
	k, err := kafka.NewKafka(&kafka.Config{
		Brokers: []string{"localhost:9092"},
	})

	assert.Nil(t, err)
	ctx := context.Background()

	producer, err := CreateProducer(k,
		SetContext(ctx),
		SetTopic("my-topic"),
		SetPartitionerMode(RoundRobin),
		SetProduceMode(SyncMode),
		SetRequireAsks())

	assert.Nil(t, err)
	producer.AddHook(&testHook{})

	now := time.Now().UTC()

	m, err := producer.Produce(ctx, &Message{
		Topic:     "my-topic",
		Key:       "1",
		Value:     "1",
		Partition: 0,
		Offset:    2,
		Timestamp: now,
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-topic", m.Topic)
	assert.Equal(t, "1", m.Key)
	assert.Equal(t, "1", m.Value)
	assert.Equal(t, int32(0), m.Partition)
	assert.NotEqual(t, int64(999), m.Offset)
	assert.Equal(t, now.String(), m.Timestamp.String())

	t.Logf("%v", producer.GetStats())
	err = producer.Close()
	assert.Nil(t, err)
}

func TestBulkProduce(t *testing.T) {
	logger.InitLogger(logger.DebugLevel)
	k, err := kafka.NewKafka(&kafka.Config{
		Brokers: []string{"localhost:9092"},
	})

	assert.Nil(t, err)
	ctx := context.Background()

	producer, err := CreateProducer(k,
		SetContext(ctx),
		SetTopic("my-topic"),
		SetPartitionerMode(RoundRobin),
		SetProduceMode(SyncMode),
		SetRequireAsks())

	assert.Nil(t, err)
	producer.AddHook(&testHook{})

	now := time.Now().UTC()

	for i := 0; i < 100; i++ {
		_, err := producer.Produce(ctx, &Message{
			Topic:     "my-topic",
			Value:     fmt.Sprintf("%d", i),
			Timestamp: now,
		})
		assert.Nil(t, err)
	}

	t.Logf("%v", producer.GetStats())
	err = producer.Close()
	assert.Nil(t, err)
}

func TestAsyncProducer(t *testing.T) {
	logger.InitLogger(logger.DebugLevel)
	k, err := kafka.NewKafka(&kafka.Config{
		Brokers: []string{"localhost:9092"},
	})

	assert.Nil(t, err)
	ctx := context.Background()

	producer, err := CreateProducer(k,
		SetContext(ctx),
		SetTopic("my-topic-2"),
		SetPartitionerMode(RoundRobin),
		SetProduceMode(AsyncMode),
		SetRequireAsks())

	assert.Nil(t, err)

	producer.AddHook(&testHook{})

	now := time.Now().UTC()

	m, err := producer.Produce(ctx, &Message{
		Topic:     "my-topic-2",
		Key:       "1",
		Value:     "1",
		Partition: 0,
		Offset:    2,
		Timestamp: now,
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-topic-2", m.Topic)
	assert.Equal(t, "1", m.Key)
	assert.Equal(t, "1", m.Value)
	assert.Equal(t, int32(0), m.Partition)
	assert.NotEqual(t, int64(999), m.Offset)
	assert.Equal(t, now.String(), m.Timestamp.String())

	t.Logf("%v", producer.GetStats())
	err = producer.Close()
	assert.Nil(t, err)
}
