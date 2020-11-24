package produce

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

type Message struct {
	Topic     string      `json:"topic"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Partition int32       `json:"partition"`
	Offset    int64       `json:"offset"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewMessage(key string, value interface{}) *Message {
	return &Message{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UTC(),
	}
}

func NewSimpleMessage(value interface{}) *Message {
	return &Message{
		Value:     value,
		Timestamp: time.Now().UTC(),
	}
}

func (m *Message) ToProducerMessage() (*sarama.ProducerMessage, error) {
	b, err := json.Marshal(m.Value)
	if err != nil {
		return nil, fmt.Errorf("marshall produce message has error: %v", err.Error())
	}

	return &sarama.ProducerMessage{
		Topic:     m.Topic,
		Key:       sarama.StringEncoder(m.Key),
		Value:     sarama.ByteEncoder(b),
		Timestamp: m.Timestamp,
		Partition: m.Partition,
		Offset:    m.Offset,
	}, nil
}
