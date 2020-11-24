package produce

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"

	"github.com/1infras/go-kit/driver/kafka"
	"github.com/1infras/go-kit/lib/hook/common"
	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/util"
)

type ProducerMode int
type PartitionerMode int

const (
	SyncMode ProducerMode = iota
	AsyncMode
)

const (
	Random PartitionerMode = iota
	RoundRobin
	Hash
)

type Produce interface {
	Produce(ctx context.Context, message *Message) (*Message, error)
	Close() error
	AddHook(hook common.HookProcess)
	Report(ctx context.Context) string
}

type ProducerOptionFunc func(*Producer) error

type Producer struct {
	context    context.Context
	cancelFunc context.CancelFunc

	produceMode     ProducerMode
	partitionerMode PartitionerMode
	requireAsks     bool

	topic string

	syncProducer  *syncProducer
	asyncProducer *asyncProducer

	stat *stat
	hook *common.Hook
}

func (_this *Producer) Close() error {
	if _this.produceMode == AsyncMode {
		return _this.asyncProducer.close()
	}
	return _this.syncProducer.close()
}

func (_this *Producer) AddHook(hook common.HookProcess) {
	_this.hook.AddHook(hook)
}

func (_this *Producer) Produce(ctx context.Context, message *Message) (m *Message, err error) {
	_this.hook.Process(ctx, func() {
		m, err = _this.produce(message)
	}, "produce")

	return
}

func (_this *Producer) Report(ctx context.Context) (result string) {
	_this.hook.Process(ctx, func() {
		result = _this.report()
	}, "report")

	return
}

// CreateProducer
func CreateProducer(client *kafka.Kafka, options ...ProducerOptionFunc) (Produce, error) {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Producer{
		context:         ctx,
		cancelFunc:      cancel,
		produceMode:     SyncMode,
		partitionerMode: RoundRobin,
		requireAsks:     true,
		stat: &stat{
			timeStart: time.Now().Unix(),
		},
		hook: &common.Hook{},
	}

	for _, option := range options {
		if err := option(p); err != nil {
			return nil, err
		}
	}

	if p.topic == "" {
		return nil, fmt.Errorf("kafka topic must be defined")
	}

	cfg := sarama.NewConfig()
	cfg.Version = client.Version
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Producer.CompressionLevel = 6

	if client.TLS != nil {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = client.TLS
	}

	switch p.partitionerMode {
	case Random:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	case RoundRobin:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case Hash:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	}

	if p.requireAsks {
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	}

	if p.produceMode == SyncMode {
		sp, err := sarama.NewSyncProducer(client.Brokers, cfg)
		if err != nil {
			return nil, fmt.Errorf("create sync producer has error: %v", err)
		}
		p.syncProducer = &syncProducer{
			topic: p.topic,
			p:     sp,
		}
	} else {
		ap, err := sarama.NewAsyncProducer(client.Brokers, cfg)
		if err != nil {
			return nil, fmt.Errorf("create async producer has error: %v", err)
		}
		p.asyncProducer = &asyncProducer{
			topic: p.topic,
			p:     ap,
		}

		p.sync()
	}

	return p, nil
}

func SetProduceMode(mode ProducerMode) ProducerOptionFunc {
	return func(p *Producer) error {
		p.produceMode = mode
		return nil
	}
}

func SetContext(ctx context.Context) ProducerOptionFunc {
	return func(p *Producer) error {
		if ctx == nil {
			return fmt.Errorf("context must not be empty")
		}
		ctx, cancel := util.GetContextWithCancel(ctx)
		p.context = ctx
		p.cancelFunc = cancel
		return nil
	}
}

func SetPartitionerMode(mode PartitionerMode) ProducerOptionFunc {
	return func(p *Producer) error {
		p.partitionerMode = mode
		return nil
	}
}

func SetTopic(topic string) ProducerOptionFunc {
	return func(p *Producer) error {
		if topic == "" {
			return fmt.Errorf("topic must be defined")
		}
		p.topic = topic
		return nil
	}
}

func SetRequireAsks() ProducerOptionFunc {
	return func(p *Producer) error {
		p.requireAsks = true
		return nil
	}
}

func (_this *Producer) produce(message *Message) (*Message, error) {
	if message == nil {
		return nil, fmt.Errorf("message must not be empty")
	}

	var (
		m   *Message
		err error
	)

	defer func() {
		atomic.AddUint32(&_this.stat.totalOperations, 1)

		b, err := json.Marshal(message.Value)
		if err == nil {
			atomic.AddInt64(&_this.stat.totalReceivedBytes, int64(len(b)))
		}
	}()

	if _this.produceMode == AsyncMode {
		m, err = _this.asyncProducer.produce(message)
	} else {
		m, err = _this.syncProducer.produce(message)
		if err == nil {
			atomic.AddUint32(&_this.stat.totalSuccesses, 1)
		} else {
			atomic.AddUint32(&_this.stat.totalErrors, 1)
		}
	}

	return m, err
}

func (_this *Producer) report() string {
	duration := time.Since(time.Unix(_this.stat.timeStart, 0)).Seconds()

	return fmt.Sprintf(`
		#Producer Stat
		Total ops: %v,
		Total sucesses ops: %v,
		Total errors ops: %v,
		Total received bytes: %v,
		Total ops per second: %.2f,
		Total sucesses ops per second: %v,
		Total errors ops per second: %v`,
		_this.stat.totalOperations,
		_this.stat.totalSuccesses,
		_this.stat.totalErrors,
		_this.stat.totalReceivedBytes,
		float64(_this.stat.totalOperations*1.0)/duration,
		float64(_this.stat.totalSuccesses*1.0)/duration,
		float64(_this.stat.totalErrors*1.0)/duration)
}

func (_this *Producer) sync() {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-_this.asyncProducer.p.Successes():
				atomic.AddUint32(&_this.stat.totalSuccesses, 1)
			case err := <-_this.asyncProducer.p.Errors():
				atomic.AddUint32(&_this.stat.totalErrors, 1)
				msg, _ := err.Msg.Value.Encode()
				key, _ := err.Msg.Key.Encode()
				logger.Error("produce message to kafka has failed",
					zap.String("error", err.Error()),
					zap.String("topic", err.Msg.Topic),
					zap.Int64("offset", err.Msg.Offset),
					zap.Int32("partition", err.Msg.Partition),
					zap.String("key", string(key)),
					zap.String("value", string(msg)))
			}
		}
	}(_this.context)
}
