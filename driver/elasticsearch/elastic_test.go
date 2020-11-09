package elasticsearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewElasticClient(t *testing.T) {
	client, err := NewElasticClient(&Connection{
		URL:         "http://localhost:9200",
		Secure:      false,
		APIKey:      "",
		EnableSniff: true,
		MaxRetries:  3,
		RetryAfter:  3,
	})

	assert.Nil(t, err)
	assert.Equal(t, true, client.IsRunning())
}
