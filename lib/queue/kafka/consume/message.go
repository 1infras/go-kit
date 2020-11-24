package consume

import "github.com/Shopify/sarama"

type ConsumerSessionMessage struct {
	Session sarama.ConsumerGroupSession
	Message *sarama.ConsumerMessage
}

type ConsumeStatus int

const (
	Consumed ConsumeStatus = iota
	Dispeard
	Error
	Retry
)