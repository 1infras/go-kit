package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	"go.elastic.co/apm/module/apmgoredis"
)

const (
	//RedisSingle -
	RedisSingle = "single"
	//RedisSentinel -
	RedisSentinel = "sentinel"
	//RedisCluster -
	RedisCluster = "cluster"
	//DefaultRedisAddress -
	DefaultRedisAddress = "localhost:6379"
	//DefaultMaxRetries -
	DefaultMaxRetries = 3
	//DefaultPoolSize -
	DefaultPoolSize = 100
)

//Connection - Config connection to Redis
type Connection struct {
	//MasterName - Master name of SentinelCluster
	MasterName string `json:"master_name"`
	//Address - Address of redis single
	Address string `json:"address"`
	//Addresses - Addresses of redis cluster or sentinel
	Addresses []string `json:"addresses"`
	//Password - Optional
	Password string `json:"password"`
	//DB - Default is 0
	DB int `json:"db"`
	//Max retries - Default is 3
	MaxRetries int `json:"max_retries"`
	// PoolSize - Default is 100
	PoolSize int `json:"pool_size"`
}

//ConfigWithDefault - Get Connection config with default
func ConfigWithDefault(c *Connection) {
	if c.MaxRetries <= 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.PoolSize <= 0 {
		c.PoolSize = DefaultPoolSize
	}

	if c.DB <= 0 {
		c.DB = 0
	}
}

//NewUniversalRedisClient - New a redis client base on configuration
//If you want to use Sentinel, set MasterName
//If you want to use Cluster, Set Addresses more than one string
//If you want to use Single, Set Addresses or Address with a string
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

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

//NewClusterRedisClient - New a redis client with cluster mode
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

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

//NewSentinelRedisClient - New a redis client with sentinel mode
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

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

//NewSingleRedisClient - New a redis client with single mode
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

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

//NewDefaultRedisClient - New a redis client with default address and single mode
func NewDefaultRedisClient() (*redis.Client, error) {
	return NewSingleRedisClient(&Connection{
		Address:    DefaultRedisAddress,
		Password:   "",
		DB:         0,
		MaxRetries: DefaultMaxRetries,
		PoolSize:   DefaultPoolSize,
	})
}

//Instrument - Wrap Redis UniversalClient with APM
//For example:
//ctx := r.Context() /r is HTTP Request Context
//client, err := NewUniversalRedisClient(&redis.Connection{Address: "localhost:6379"})
//if err != nil {panic(err)}
//client = Instrument(ctx, client)
func Instrument(ctx context.Context, client redis.UniversalClient) redis.UniversalClient {
	return apmgoredis.Wrap(client).WithContext(ctx)
}
