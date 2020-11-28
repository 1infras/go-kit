package kafka

import (
	"os"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

func TestNewKafka(t *testing.T) {
	err := os.Setenv("KAFKA_BROKERS", "broker1:9200,broker2:9200")
	assert.Nil(t, err)
	err = os.Setenv("KAFKA_TLS", "false")
	assert.Nil(t, err)
	err = os.Setenv("KAFKA_VERSION", "2.6.0")
	assert.Nil(t, err)
	k, err := NewKafka(nil)
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.NotNil(t, k.Brokers)
	assert.Equal(t, "broker1:9200", k.Brokers[0])
	assert.Equal(t, "broker2:9200", k.Brokers[1])
	assert.Nil(t, k.TLS)
	assert.Equal(t, sarama.V2_6_0_0, k.Version)
}