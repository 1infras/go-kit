package redis

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"

	"github.com/go-redis/redis"
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

// Connection - Config connection to Redis
type Connection struct {
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

// ConfigWithDefault - Get Connection config with default
func (c *Connection) Default() {
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

func ProcessConnection() (*Connection, error) {
	var c *Connection
	err := envconfig.Process("redis", &c)
	if err != nil {
		return nil, err
	}

	// If connection wasn't found in envconfig then try with viper
	if c == nil {
		c = &Connection{
			MasterName: viper.GetString("redis.master_name"),
			Address:    viper.GetString("redis.address"),
			Addresses:  viper.GetStringSlice("redis.addresses"),
			Password:   viper.GetString("redis.password"),
			DB:         viper.GetInt("redis.db"),
			MaxRetries: viper.GetInt("redis.max_retries"),
			PoolSize:   viper.GetInt("redis.pool_size"),
			RetryAfter: viper.GetDuration("redis.retry_after"),
		}
	}

	c.Default()

	return c, nil
}

// NewUniversalRedisClient - New a redis client base on configuration
// If you want to use Sentinel, set MasterName
// If you want to use Cluster, Set Addresses more than one string
// If you want to use Single, Set Addresses or Address with a string
func NewUniversalRedisClient(c *Connection) (redis.UniversalClient, error) {
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

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

// NewClusterRedisClient - New a redis client with cluster mode
func NewClusterRedisClient(c *Connection) (*redis.ClusterClient, error) {
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

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

// NewSentinelRedisClient - New a redis client with sentinel mode
func NewSentinelRedisClient(c *Connection) (*redis.Client, error) {
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

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

// NewSingleRedisClient - New a redis client with single mode
func NewSingleRedisClient(c *Connection) (*redis.Client, error) {
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

	for {
		_, err := client.Ping().Result()
		if err == nil {
			break
		}
		if retries >= c.MaxRetries {
			return nil, err
		}
		retries++
		time.Sleep(c.RetryAfter)
	}

	return client, nil
}

// NewDefaultRedisClient - New a redis client with default address and single mode
func NewDefaultRedisClient() (*redis.Client, error) {
	return NewSingleRedisClient(&Connection{
		Address:    DefaultRedisAddress,
		Password:   "",
		DB:         0,
		MaxRetries: DefaultMaxRetries,
		PoolSize:   DefaultPoolSize,
	})
}

// NewDefaultRedisUniversalClient - New a redis universal client with default address and single mode
func NewDefaultRedisUniversalClient() (redis.UniversalClient, error) {
	c := &Connection{}
	c.Default()
	return NewUniversalRedisClient(c)
}
