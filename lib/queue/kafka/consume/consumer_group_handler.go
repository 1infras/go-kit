package consume

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"

	"github.com/1infras/go-kit/lib/hook/common"
)

type stream []*ConsumerSessionMessage

type consumerGroupHandler struct {
	ctx context.Context
	ready chan bool

	bufferCapability int
	task             int

	bufferStream stream
	mainStream   chan stream
	hook         *common.Hook

	ticker *time.Ticker
	lock   sync.RWMutex
}

func (_this *consumerGroupHandler) processMessage(stat *stat, fn func(message []byte) ConsumeStatus) {
	for i := 0; i < _this.task; i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case ms := <-_this.mainStream:
					for _, m := range ms {
						atomic.AddUint32(&stat.totalOperations, 1)
						atomic.AddInt64(&stat.totalReceivedBytes, int64(len(m.Message.Value)))

						var status ConsumeStatus
						_this.hook.Process(context.Background(), func() {
							status = fn(m.Message.Value)
						}, "consumer.kafka")

						switch status {
						case Consumed:
							atomic.AddUint32(&stat.totalConsumed, 1)
							m.Session.MarkMessage(m.Message, "")
						case Dispeard:
							atomic.AddUint32(&stat.totalDispeared, 1)
							m.Session.MarkMessage(m.Message, "")
						case Error:
							atomic.AddUint32(&stat.totalErrors, 1)
						case Retry:
							atomic.AddUint32(&stat.totalRetry, 1)
						}
					}
				}
			}
		}(_this.ctx)
	}
}

func (_this *consumerGroupHandler) flushBuffer() {
	_this.lock.Lock()
	defer _this.lock.Unlock()
	if len(_this.bufferStream) > 0 {
		_this.mainStream <- _this.bufferStream
		_this.bufferStream = make(stream, 0, _this.bufferCapability)
	}
}

func (_this *consumerGroupHandler) catchMessage(m *ConsumerSessionMessage) {
	_this.lock.Lock()
	defer _this.lock.Unlock()
	_this.bufferStream = append(_this.bufferStream, m)
	if len(_this.bufferStream) > _this.bufferCapability {
		_this.flushBuffer()
	}
}

func (_this *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(_this.ready)
	return nil
}

func (_this *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (_this *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c := claim.Messages()
	for {
		select {
		case <-_this.ctx.Done():
			return nil
		case m, ok := <-c:
			if ok {
				_this.catchMessage(&ConsumerSessionMessage{
					Session: session,
					Message: m,
				})
			}
		case <-_this.ticker.C:
			_this.flushBuffer()
		}
	}
}