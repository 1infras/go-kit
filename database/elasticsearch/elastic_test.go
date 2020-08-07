package elasticsearch

import (
	"testing"
)

func TestNewDefaultElasticClient(t *testing.T) {
	_, err := NewElasticClient(&Connection{URL: "http://127.0.0.1:9200"})
	if err != nil {
		t.Fatalf("elastic must be connected, error was throwed: %v", err)
	}
}
