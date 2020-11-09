package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"

	"github.com/1infras/go-kit/logger"
)

// NewConsumerConfig -
func NewConsumerConfig(conn *Connection) *sarama.Config {
	c := sarama.NewConfig()

	if conn.TLS != nil {
		c.Net.TLS.Enable = true
		c.Net.TLS.Config = conn.TLS
	}

	c.Consumer.Offsets.Initial = sarama.OffsetNewest
	c.Consumer.Return.Errors = true

	return c
}

// CreateConsumer -
func CreateConsumer(conn *Connection) (sarama.Consumer, error) {
	if conn == nil {
		return CreateConsumerFromDefaultConnection()
	}

	c := NewConsumerConfig(conn)
	return sarama.NewConsumer(conn.Brokers, c)
}

// CreateConsumerGroup -
func CreateConsumerGroup(conn *Connection, group string) (sarama.ConsumerGroup, error) {
	if group == "" {
		return nil, fmt.Errorf("group must be defined")
	}

	if conn == nil {
		return CreateConsumerGroupFromDefaultConnection(group)
	}

	c := NewConsumerConfig(conn)
	return sarama.NewConsumerGroup(conn.Brokers, group, c)
}

// CreateConsumerFromDefaultConnection -
func CreateConsumerFromDefaultConnection() (sarama.Consumer, error) {
	c, err := NewDefaultKafkaConnection(nil)
	if err != nil {
		return nil, err
	}

	return CreateConsumer(c)
}

// CreateConsumerGroupFromDefaultConnection -
func CreateConsumerGroupFromDefaultConnection(group string) (sarama.ConsumerGroup, error) {
	if group == "" {
		return nil, fmt.Errorf("group must be defined")
	}

	c, err := NewDefaultKafkaConnection(nil)
	if err != nil {
		return nil, err
	}

	return CreateConsumerGroup(c, group)
}

// Consume -
func Consume(consumer sarama.Consumer, topic string, fn func(msg *sarama.ConsumerMessage)) error {
	if topic == "" {
		return fmt.Errorf("topic must be defined")
	}

	var (
		sigChan = make(chan os.Signal, 1)
		mc      = make(chan *sarama.ConsumerMessage, 256)
	)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, p := range partitions {
		pc, err := consumer.ConsumePartition(topic, p, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for m := range pc.Messages() {
				mc <- m
			}
		}(pc)
	}

	run := true
	for run == true {
		select {
		case <-sigChan:
			run = false
		case message := <-mc:
			fn(message)
		}
	}

	close(sigChan)
	logger.Infof("Closing consumer")
	return consumer.Close()
}

// TODO: Write ConsumeGroup
// ConsumeGroup -
func ConsumeGroup(consumerGroup sarama.ConsumerGroup, topic string, fn func(msg *sarama.ConsumerMessage)) error {
	return nil
}
