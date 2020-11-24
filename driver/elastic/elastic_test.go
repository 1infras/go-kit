package elastic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewElasticClient(t *testing.T) {
	client, err := NewElasticClient(&Config{
		URL:         "http://localhost:9200",
		Secure:      false,
		APIKey:      "",
		MaxRetries:  3,
		RetryAfter:  3,
	})

	assert.Nil(t, err)
	assert.NotNil(t, client)

	res1, err := client.Cat.Health()
	assert.Nil(t, err)
	assert.NotNil(t, res1)

	t.Logf("%s", res1)

	res2, err := client.Index("test-index", strings.NewReader(`{"title":"Test"}`), client.Index.WithDocumentID("1"))
	assert.Nil(t, err)
	assert.NotNil(t, res2)

	t.Logf("%s", res2)
}
