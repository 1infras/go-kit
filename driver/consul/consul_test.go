package consul

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessConfig(t *testing.T) {
	os.Setenv("CONSUL_ENDPOINT", "http://localhost:8500")
	os.Setenv("CONSUL_TOKEN", "123456")
	os.Setenv("CONSUL_DATACENTER", "dc1")
	os.Setenv("CONSUL_NAMESPACE", "ns1")

	cfg, err := ProcessConfig(nil)
	assert.Nil(t, err)
	assert.Equal(t, "http://localhost:8500", cfg.Address)
	assert.Equal(t, "123456", cfg.Token)
	assert.Equal(t, "dc1", cfg.Datacenter)
	assert.Equal(t, "ns1", cfg.Namespace)
}

func TestConsul(t *testing.T) {
	cfg := &Config{Endpoint: "http://localhost:8500"}
	consul, err := NewConsul(cfg)
	assert.Nil(t, err)

	sample := struct {
		Foo string `json:"foo"`
	}{
		Foo: "bar",
	}

	b, err := json.Marshal(sample)
	assert.Nil(t, err)

	err = consul.PutKV("foo", b)
	assert.Nil(t, err)

	k, err := consul.GetKeys("foo")
	assert.Nil(t, err)

	if len(k) > 1 {
		t.Fatalf("Expected: 1, Actual: %v", len(k))
	}

	assert.Equal(t, "foo", k[0])

	v, err := consul.GetKV("foo")
	assert.Nil(t, err)

	s := struct {
		Foo string `json:"foo"`
	}{}

	err = json.Unmarshal(v, &s)
	assert.Nil(t, err)
	assert.Equal(t, "bar", s.Foo)

	_, err = consul.DeleteKV("foo")
	assert.Nil(t, err)

	_, err = consul.GetKV("foo")
	assert.NotNil(t, err)
}
