package kafka

import (
	"crypto/tls"
	"fmt"
	"github.com/1infras/go-kit/util/cert_utils"
	"github.com/1infras/go-kit/util/file_utils"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

//Connection of Kafka Cluster/Standalone
type Connection struct {
	Brokers []string    `json:"brokers"` //The list of kafka brokers
	TLS     *tls.Config `json:"tls"`     //SSL configuration
}

//NewDefaultKafkaConnection with settings from viper
func NewDefaultKafkaConnection() (*Connection, error) {
	servers := viper.GetStringSlice("kafka.servers")
	if len(servers) == 0 {
		return nil, fmt.Errorf("kafka servers must be defined")
	}

	c := &Connection{
		Brokers: servers,
	}

	if viper.GetBool("kafka.tls") {
		tlsClientCert, err := file_utils.GetAbsolutelyLocalFilePath("kafka.tls_client_cert")
		if err != nil {
			return nil, err
		}

		tlsClientKey, err := file_utils.GetAbsolutelyLocalFilePath("kafka.tls_client_key")
		if err != nil {
			return nil, err
		}

		tlsClientCA := viper.GetString("kafka.tls_client_ca")
		tlsSkipVerify := viper.GetBool("kafka.tls_skip_verify")

		if tlsClientCA != "" {
			tlsClientCA, err = file_utils.GetAbsolutelyLocalFilePath("kafka.tls_client_ca")
			if err != nil {
				return nil, err
			}
		}

		tlsConfig, err := cert_utils.NewTLS(tlsClientCert, tlsClientKey, tlsClientCA, tlsSkipVerify)
		if err != nil {
			return nil, err
		}

		c.TLS = tlsConfig
	}

	return c, nil
}

//NewKafkaConfig
func NewKafkaConfig(conn *Connection) *sarama.Config {
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

//CreateAsyncProducer
func CreateAsyncProducer(conn *Connection) (sarama.AsyncProducer, error) {
	if conn == nil {
		return CreateAsyncProducerFromDefaultConnection()
	}

	c := NewKafkaConfig(conn)
	return sarama.NewAsyncProducer(conn.Brokers, c)
}

//CreateSyncProducer
func CreateSyncProducer(conn *Connection) (sarama.SyncProducer, error) {
	if conn == nil {
		return CreateSyncProducerFromDefaultConnection()
	}

	c := NewKafkaConfig(conn)
	return sarama.NewSyncProducer(conn.Brokers, c)
}

//CreateAsyncProducerFromDefaultConnection
func CreateAsyncProducerFromDefaultConnection() (sarama.AsyncProducer, error) {
	c, err := NewDefaultKafkaConnection()
	if err != nil {
		return nil, err
	}
	return CreateAsyncProducer(c)
}

//CreateSyncProducerFromDefaultConnection
func CreateSyncProducerFromDefaultConnection() (sarama.SyncProducer, error) {
	c, err := NewDefaultKafkaConnection()
	if err != nil {
		return nil, err
	}
	return CreateSyncProducer(c)
}
