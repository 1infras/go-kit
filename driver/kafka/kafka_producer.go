package kafka

import "github.com/Shopify/sarama"

// NewProducerConfig
func NewProducerConfig(conn *Connection) *sarama.Config {
	c := sarama.NewConfig()

	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true

	if conn.TLS != nil {
		c.Net.TLS.Enable = true
		c.Net.TLS.Config = conn.TLS
	}

	c.Producer.Partitioner = sarama.NewRandomPartitioner

	return c
}

// CreateAsyncProducer
func CreateAsyncProducer(conn *Connection) (sarama.AsyncProducer, error) {
	if conn == nil {
		return CreateAsyncProducerFromDefaultConnection()
	}

	c := NewProducerConfig(conn)

	return sarama.NewAsyncProducer(conn.Brokers, c)
}

// CreateSyncProducer
func CreateSyncProducer(conn *Connection) (sarama.SyncProducer, error) {
	if conn == nil {
		return CreateSyncProducerFromDefaultConnection()
	}

	c := NewProducerConfig(conn)

	return sarama.NewSyncProducer(conn.Brokers, c)
}

// CreateAsyncProducerFromDefaultConnection
func CreateAsyncProducerFromDefaultConnection() (sarama.AsyncProducer, error) {
	c, err := NewDefaultKafkaConnection(nil)
	if err != nil {
		return nil, err
	}

	return CreateAsyncProducer(c)
}

// CreateSyncProducerFromDefaultConnection
func CreateSyncProducerFromDefaultConnection() (sarama.SyncProducer, error) {
	c, err := NewDefaultKafkaConnection(nil)
	if err != nil {
		return nil, err
	}

	return CreateSyncProducer(c)
}
