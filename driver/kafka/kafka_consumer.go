package kafka

import (
	"github.com/1infras/go-kit/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

//KafkaConsumerConfigMap --
type ConsumerConfigMap struct {
	Servers            string `json:"servers"`
	Topic              string `json:"topic"`
	TLS                bool   `json:"tls"`
	TrustStorePath     string `json:"trust_store_path"`
	TrustStorePassword string `json:"trust_store_password"`
	KeyStorePath       string `json:"key_store_path"`
	KeyStorePassword   string `json:"key_store_password"`
	Offset             string `json:"offset"`
	SessionTimeoutMs   int    `json:"session_timeout_ms"`
	ConsumerGroupID    string `json:"consumer_group_id"`
}

//LoadConsumerConfigMap --
func LoadConsumerConfigMap() *ConsumerConfigMap {
	return &ConsumerConfigMap{
		Servers:            viper.GetString("kafka.servers"),
		Topic:              viper.GetString("kafka.topic"),
		TLS:                viper.GetBool("kafka.tls"),
		TrustStorePath:     viper.GetString("kafka.truststore_path"),
		TrustStorePassword: viper.GetString("kafka.truststore_password"),
		KeyStorePath:       viper.GetString("kafka.keystore_path"),
		KeyStorePassword:   viper.GetString("kafka.keystore_password"),
		Offset:             viper.GetString("kafka.offset"),
		SessionTimeoutMs:   viper.GetInt("kafka.session_timeout_ms"),
		ConsumerGroupID:    viper.GetString("kafka.consumer_group_id"),
	}
}

//KafkaConsumer --
type Consumer struct {
	Consumer *kafka.Consumer
	Topics   []string
}

//ToConfigMap --
func (c *ConsumerConfigMap) ToConfigMap() *kafka.ConfigMap {
	if c.Offset == "" {
		c.Offset = "latest"
	}

	if c.ConsumerGroupID == "" {
		c.ConsumerGroupID = "flink-kafka-consumer-group"
	}

	if c.SessionTimeoutMs <= 0 {
		c.SessionTimeoutMs = 60000
	}

	cm := &kafka.ConfigMap{
		"bootstrap.servers":               c.Servers,
		"broker.address.family":           "v4", //Avoid connecting to IPv6 brokers
		"group.id":                        c.ConsumerGroupID,
		"session.timeout.ms":              c.SessionTimeoutMs,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               c.Offset,
	}

	if c.TLS {
		cm.SetKey("security.protocol", "SSL")
		cm.SetKey("ssl.truststore.location", c.TrustStorePath)
		cm.SetKey("ssl.truststore.password", c.TrustStorePassword)
		cm.SetKey("ssl.keystore.location", c.KeyStorePath)
		cm.SetKey("ssl.keystore.password", c.KeyStorePassword)
	}

	return cm
}

//NewConsumer --
func NewConsumer(configMap *ConsumerConfigMap) (*Consumer, error) {
	c, err := kafka.NewConsumer(configMap.ToConfigMap())
	if err != nil {
		return nil, err
	}

	return &Consumer{
		Consumer: c,
		Topics:   []string{configMap.Topic}}, nil
}

//Subscribes --
func (c *Consumer) Subscribes(fn func(message *kafka.Message)) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	err := c.Consumer.SubscribeTopics(c.Topics, nil)
	if err != nil {
		return err
	}

	logger.Infof("Starting consume topic: %v", strings.Join(c.Topics, ", "))

	for {
		select {
		case <- sigChan:
			logger.Infof("Closing consumer")
			err := c.Consumer.Close()
			if err != nil {
				return err
			}
			close(sigChan)
			return nil
		case event := <-c.Consumer.Events():
			switch e := event.(type) {
			case kafka.AssignedPartitions:
				err := c.Consumer.Assign(e.Partitions)
				if err != nil {
					return err
				}
				logger.Infof("Assigned Partitions %v\n", e)
			case kafka.RevokedPartitions:
				err := c.Consumer.Unassign()
				if err != nil {
					return err
				}
				logger.Infof("Revoked Partitions %v\n", e)
			case kafka.PartitionEOF:
			case kafka.Error:
				return e
			case *kafka.Message:
				fn(e)
			}
		}
	}
}