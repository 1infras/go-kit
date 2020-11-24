package consume

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"

	"github.com/1infras/go-kit/driver/kafka"
	"github.com/1infras/go-kit/lib/hook/common"
	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/util"
)

type InitialOffsetMode int
type BalanceStrategyMode int
type ConsumeHandler func(message []byte) ConsumeStatus

const (
	Oldest InitialOffsetMode = iota
	Newest
)

const (
	Sticky BalanceStrategyMode = iota
	RoundRobin
	Range
)

type Consume interface {
	SetConsumeHandler(ConsumeHandler)
	Run()
	Report() string
	AddHook(hook common.HookProcess)
}

type ConsumerOptionFunc func(*Consumer) error

type Consumer struct {
	context    context.Context
	cancelFunc context.CancelFunc

	initialOffsetMode   InitialOffsetMode
	balanceStrategyMode BalanceStrategyMode

	consumeHandler ConsumeHandler

	topic            string
	group            string
	bufferCapability int
	task             int

	consumerGroup *consumerGroup
	closeFunc     func()

	stat *stat
	hook *common.Hook
}

func (_this *Consumer) AddHook(hook common.HookProcess) {
	_this.hook.AddHook(hook)
}

func (_this *Consumer) Report() string {
	return _this.report()
}

func (_this *Consumer) SetConsumeHandler(h ConsumeHandler) {
	_this.consumeHandler = h
}

func CreateConsumer(client *kafka.Kafka, options ...ConsumerOptionFunc) (Consume, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Consumer{
		context:             ctx,
		cancelFunc:          cancel,
		initialOffsetMode:   Oldest,
		balanceStrategyMode: RoundRobin,
		bufferCapability:    250,
		task:                2,
		stat:                &stat{timeStart: time.Now().Unix()},
		hook:                &common.Hook{},
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	if c.topic == "" {
		return nil, fmt.Errorf("kafka topic must not be empty")
	}

	if c.group == "" {
		return nil, fmt.Errorf("kafka consumer group must not empty")
	}

	cfg := sarama.NewConfig()
	cfg.Version = client.Version
	cfg.Consumer.Return.Errors = true

	if client.TLS != nil {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = client.TLS
	}

	switch c.balanceStrategyMode {
	case Sticky:
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case RoundRobin:
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case Range:
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	}

	switch c.initialOffsetMode {
	case Newest:
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	case Oldest:
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	cg, err := sarama.NewConsumerGroup(client.Brokers, c.group, cfg)
	if err != nil {
		return nil, fmt.Errorf("create consumer group has error: %v", err)
	}

	c.consumerGroup = &consumerGroup{cg: cg}

	return c, nil
}

func SetContext(ctx context.Context) ConsumerOptionFunc {
	return func(c *Consumer) error {
		if ctx == nil {
			return fmt.Errorf("context must not be empty")
		}
		ctx, cancel := util.GetContextWithCancel(ctx)
		c.context = ctx
		c.cancelFunc = cancel
		return nil
	}
}

func SetInitialOffsetMode(mode InitialOffsetMode) ConsumerOptionFunc {
	return func(c *Consumer) error {
		c.initialOffsetMode = mode
		return nil
	}
}

func SetBalanceStrategyMode(mode BalanceStrategyMode) ConsumerOptionFunc {
	return func(c *Consumer) error {
		c.balanceStrategyMode = mode
		return nil
	}
}

func SetTopic(topic string) ConsumerOptionFunc {
	return func(c *Consumer) error {
		c.topic = topic
		return nil
	}
}

func SetGroup(group string) ConsumerOptionFunc {
	return func(c *Consumer) error {
		c.group = group
		return nil
	}
}

func SetClose(fn func()) ConsumerOptionFunc {
	return func(c *Consumer) error {
		c.closeFunc = fn
		return nil
	}
}

func (_this *Consumer) getConsumeGroupHandler() *consumerGroupHandler {
	return &consumerGroupHandler{
		ctx:              _this.context,
		bufferCapability: _this.bufferCapability,
		task:             _this.task,
		bufferStream:     make(stream, 0, _this.bufferCapability),
		mainStream:       make(chan stream, 1000),
		ticker:           time.NewTicker(1 * time.Second),
		hook:             _this.hook,
		ready:            make(chan bool),
	}
}

func (_this *Consumer) close() {
	_this.closeFunc()
	if err := _this.consumerGroup.close(); err != nil {
		logger.Errorf("close consumer group has error: %v", err.Error())
	}
	logger.Info("Consumer has closed")
}

func (_this *Consumer) Run() {
	// Get consumer group handler
	handler := _this.getConsumeGroupHandler()
	// Process messages have consumed
	handler.processMessage(_this.stat, _this.consumeHandler)

	ctx, cancel := context.WithCancel(_this.context)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err := _this.consumerGroup.cg.Consume(ctx, []string{_this.topic}, handler)
			if err != nil {
				if err != sarama.ErrClosedConsumerGroup {
					logger.Error("consumer run has error", zap.String("error", err.Error()))
					os.Exit(1)
				} else {
					logger.Info("consumer has gracefully closed")
				}
			}

			if ctx.Err() != nil {
				return
			}

			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready
	logger.Info("Consumer up and running!")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		logger.Info("terminating: context canceled")
	case <-sigterm:
		logger.Info("terminating: via signal")
	}
	cancel()
	wg.Wait()
	_this.close()
}

func (_this *Consumer) report() string {
	duration := time.Since(time.Unix(_this.stat.timeStart, 0)).Seconds()

	return fmt.Sprintf(`
		#Producer Stat
		Total ops: %v,
		Total consumed ops: %v,
		Total dispeared ops: %v,
		Total errors ops: %v,
		Total retry ops: %v,
		Total received bytes: %v,
		Total ops per second: %.2f,
		Total consumed ops per second: %v,
		Total errors ops per second: %v`,
		_this.stat.totalOperations,
		_this.stat.totalConsumed,
		_this.stat.totalDispeared,
		_this.stat.totalErrors,
		_this.stat.totalRetry,
		_this.stat.totalReceivedBytes,
		float64(_this.stat.totalOperations*1.0)/duration,
		float64(_this.stat.totalConsumed*1.0)/duration,
		float64(_this.stat.totalErrors*1.0)/duration)
}
