package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"

	"github.com/1infras/go-kit/tracing"
)

const (
	// DefaultRedisAddress -
	DefaultRedisAddress = "localhost:6379"
	// DefaultMaxRetries -
	DefaultMaxRetries = 3
	// DefaultPoolSize -
	DefaultPoolSize = 100
	// DefaultRetryAfter
	DefaultRetryAfter = 5 * time.Second
)

// Config - Config connection to Redis
type Config struct {
	// MasterName - Master name of SentinelCluster
	MasterName string `mapstructure:"master_name" envconfig:"REDIS_MASTER_NAME"`
	// Address - Address of redis single
	Address string `mapstructure:"address" envconfig:"REDIS_ADDRESS"`
	// Addresses - Addresses of redis cluster or sentinel
	Addresses []string `mapstructure:"addresses" envconfig:"REDIS_ADDRESSES"`
	// Password - Optional
	Password string `mapstructure:"password" envconfig:"REDIS_PASSWORD"`
	// DB - Default is 0
	DB int `mapstructure:"db" envconfig:"REDIS_DB"`
	// Max retries - Default is 3
	MaxRetries int `mapstructure:"max_retries" envconfig:"REDIS_MAX_RETRIES"`
	// PoolSize - Default is 100
	PoolSize int `mapstructure:"pool_size" envconfig:"REDIS_POOL_SIZE"`
	// RetryAfter - Default is 5 seconds
	RetryAfter time.Duration `mapstructure:"retry_after" envconfig:"REDIS_RETRY_AFTER"`
}

// ConfigWithDefault - Get Config config with default
func (c *Config) Default() {
	if c.Address == "" && len(c.Addresses) == 0 {
		c.Address = DefaultRedisAddress
		c.Addresses = []string{DefaultRedisAddress}
	}

	if c.MaxRetries <= 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.PoolSize <= 0 {
		c.PoolSize = DefaultPoolSize
	}

	if c.DB <= 0 {
		c.DB = 0
	}

	if c.RetryAfter <= 0 {
		c.RetryAfter = DefaultRetryAfter
	}
}

func ProcessConfig(c *Config) (*Config, error) {
	if c == nil {
		c = &Config{}
		err := envconfig.Process("redis", c)
		if err != nil {
			return nil, err
		}
	}

	c.Default()

	return c, nil
}

// NewUniversalRedisClient - New a redis client base on configuration
// If you want to use Sentinel, set MasterName
// If you want to use Cluster, Set Addresses more than one string
// If you want to use Single, Set Addresses or Address with a string
func NewUniversalRedisClient(c *Config) (redis.UniversalClient, error) {
	c, err := ProcessConfig(c)
	if err != nil {
		return nil, err
	}

	if c.Address == "" && len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	options := &redis.UniversalOptions{
		Addrs:      c.Addresses,
		Password:   c.Password,
		DB:         c.DB,
		MasterName: c.MasterName,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	}

	if len(options.Addrs) == 0 {
		options.Addrs = []string{c.Address}
	}

	client := redis.NewUniversalClient(options)
	retries := 0
	ctx := context.Background()

	for {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	if tracing.Enabled {
		tracing.WrapGoRedisUniversalClient(client)
	}

	return client, nil
}

// NewClusterRedisClient
func NewClusterRedisClient(c *Config) (*redis.ClusterClient, error) {
	c, err := ProcessConfig(c)
	if err != nil {
		return nil, err
	}

	if len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:      c.Addresses,
		Password:   c.Password,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	})
	retries := 0
	ctx := context.Background()

	for {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	if tracing.Enabled {
		tracing.WrapGoRedisClusterClient(client)
	}

	return client, nil
}

// NewSentinelRedisClient
func NewSentinelRedisClient(c *Config) (*redis.Client, error) {
	c, err := ProcessConfig(c)
	if err != nil {
		return nil, err
	}

	if len(c.Addresses) == 0 {
		return nil, fmt.Errorf("address was not set")
	}

	if c.MasterName == "" {
		return nil, fmt.Errorf("master name of sentinel cluster was not set")
	}

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    c.MasterName,
		SentinelAddrs: c.Addresses,
		Password:      c.Password,
		DB:            c.DB,
		MaxRetries:    c.MaxRetries,
		PoolSize:      c.PoolSize,
	})
	retries := 0
	ctx := context.Background()

	for {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	if tracing.Enabled {
		tracing.WrapGoRedisClient(client)
	}

	return client, nil
}

// NewSingleRedisClient - New a redis client with single mode
func NewSingleRedisClient(c *Config) (*redis.Client, error) {
	c, err := ProcessConfig(c)
	if err != nil {
		return nil, err
	}

	if c.Address == "" {
		return nil, fmt.Errorf("address was not set")
	}

	client := redis.NewClient(&redis.Options{
		Addr:       c.Address,
		Password:   c.Password,
		DB:         c.DB,
		MaxRetries: c.MaxRetries,
		PoolSize:   c.PoolSize,
	})
	retries := 0
	ctx := context.Background()

	for {
		_, err := client.Ping(ctx).Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	if tracing.Enabled {
		tracing.WrapGoRedisClient(client)
	}

	return client, nil
}

// NewDefaultRedisClient
func NewDefaultRedisClient() (*redis.Client, error) {
	return NewSingleRedisClient(&Config{
		Address:    DefaultRedisAddress,
		Password:   "",
		DB:         0,
		MaxRetries: DefaultMaxRetries,
		PoolSize:   DefaultPoolSize,
	})
}

// NewDefaultRedisUniversalClient - New a redis universal client with default address and single mode
func NewDefaultRedisUniversalClient() (redis.UniversalClient, error) {
	c := &Config{}
	c.Default()
	return NewUniversalRedisClient(c)
}
