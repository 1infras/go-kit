package produce

import (
	"github.com/Shopify/sarama"
)

type syncProducer struct {
	topic string
	p     sarama.SyncProducer
}

func (_this *syncProducer) produce(msg *Message) (*Message, error) {
	if msg.Topic == "" {
		msg.Topic = _this.topic
	}

	m, err := msg.ToProducerMessage()
	if err != nil {
		return nil, err
	}

	p, o, err := _this.p.SendMessage(m)
	if err != nil {
		return nil, err
	}

	return &Message{
		Topic:     msg.Topic,
		Key:       msg.Key,
		Value:     msg.Value,
		Partition: p,
		Offset:    o,
		Timestamp: msg.Timestamp,
	}, nil
}

func (_this *syncProducer) close() error {
	return _this.p.Close()
}