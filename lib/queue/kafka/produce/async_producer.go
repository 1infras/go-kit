package produce

import (
	"github.com/Shopify/sarama"
)

type asyncProducer struct {
	topic string
	p     sarama.AsyncProducer
}

func (_this *asyncProducer) produce(msg *Message) (*Message, error) {
	if msg.Topic == "" {
		msg.Topic = _this.topic
	}

	m, err := msg.ToProducerMessage()
	if err != nil {
		return nil, err
	}

	_this.p.Input() <- m
	return msg, nil
}

func (_this *asyncProducer) close() error {
	return _this.p.Close()
}
