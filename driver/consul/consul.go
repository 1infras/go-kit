package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Endpoint   string `mapstructure:"endpoint" envconfig:"CONSUL_ENDPOINT"`
	Token      string `mapstructure:"endpoint" envconfig:"CONSUL_TOKEN"`
	Datacenter string `mapstructure:"datacenter" envconfig:"CONSUL_DATACENTER"`
	Namespace  string `mapstructure:"namespace" envconfig:"CONSUL_NAMESPACE"`
}

type kv struct {
	*api.KV
}

type KV interface {
	GetKeys(prefix string) ([]string, error)
	GetKV(key string) ([]byte, error)
	PutKV(key string, value []byte) error
	DeleteKV(key string) (bool, error)
	ExistKV(key string) (bool, error)
}

func ProcessConfig(cfg *Config) (*api.Config, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	err := envconfig.Process("consul", cfg)
	if err != nil {
		return nil, err
	}

	config := &api.Config{}
	if cfg.Endpoint != "" {
		config.Address = cfg.Endpoint
	}

	if cfg.Token != "" {
		config.Token = cfg.Token
	}

	if cfg.Datacenter != "" {
		config.Datacenter = cfg.Datacenter
	}

	if cfg.Namespace != "" {
		config.Namespace = cfg.Namespace
	}

	return config, nil
}

func NewConsul(cfg *Config) (KV, error) {
	var (
		config *api.Config
		err    error
	)

	config, err = ProcessConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	if client.KV() == nil {
		return nil, fmt.Errorf("cannot get kv with unexpected error")
	}

	return &kv{client.KV()}, nil
}

func (_this *kv) GetKeys(prefix string) ([]string, error) {
	keys, _, err := _this.Keys(prefix, "", nil)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (_this *kv) GetKV(key string) ([]byte, error) {
	pair, _, err := _this.Get(key, nil)
	if err != nil {
		return nil, err
	}

	if pair != nil {
		return pair.Value, nil
	}

	return nil, fmt.Errorf("pair of key %s is nil", key)
}

func (_this *kv) PutKV(key string, value []byte) error {
	_, err := _this.Put(&api.KVPair{
		Key:   key,
		Value: value,
	}, nil)

	return err
}

func (_this *kv) DeleteKV(key string) (bool, error) {
	_, err := _this.Delete(key, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (_this *kv) ExistKV(key string) (bool, error) {
	v, _, err := _this.Get(key, nil)
	if err != nil {
		return false, err
	}

	return v.Key != "", nil
}
