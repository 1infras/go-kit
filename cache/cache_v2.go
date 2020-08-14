package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/util"

	"github.com/1infras/go-kit/cache/codec"
	"github.com/1infras/go-kit/cache/lru"
	"github.com/go-redis/redis"
)

// OneCache --
var OneCache *OneCacheStruct

func init() {
	OneCache = &OneCacheStruct{}
}

const (
	// DefaultMaxEntries --
	DefaultMaxEntries = 1000000

	// KVType --
	KVType int = iota
	// SetType --
	SetType
	// SortedSetType --
	SortedSetType

	// BloomFilter --
	BloomFilter
)

var (
	// DefaultTimeoutInSeconds --
	DefaultTimeoutInSeconds = 30 * time.Second

	// OptimalInMemAccessTime // Follow Jeff Dean, read 1MB sequential from memory ideally is 250 micro-seconds --
	OptimalInMemAccessTime = 1 * time.Millisecond

	// OptimalNetworkAccessTime --
	OptimalNetworkAccessTime = 150 * time.Millisecond

	// ErrNotFound --
	ErrNotFound = errors.New("cache: Key is not found")
)

type cacheStat struct {
	hitCount     uint32
	missCount    uint32
	totalCount   uint32
	expiredCount uint32
	setCount     uint32
}

type cacheItem struct {
	Key      string
	Value    interface{}
	TTL      time.Duration
	DataType int
}

// Item --
type Item struct {
	value    interface{}
	rawValue []byte
}

// Bytes --
func (i *Item) Bytes() []byte {
	return i.rawValue
}

// Val --
func (i *Item) Val() interface{} {
	return i.value
}

// Int64 --
func (i *Item) Int64() int64 {
	v, _ := i.value.(int64)
	return v
}

// Int32 --
func (i *Item) Int32() int32 {
	v, _ := i.value.(int32)
	return v
}

// Int --
func (i *Item) Int() int {
	v, _ := i.value.(int)
	return v
}

// Boolean --
func (i *Item) Boolean() bool {
	v, _ := i.value.(bool)
	return v
}

// Float64 --
func (i *Item) Float64() float64 {
	v, _ := i.value.(float64)
	return v
}

// String --
func (i *Item) String() string {
	v, _ := i.value.(string)
	return v
}

// OneCacheStruct --
type OneCacheStruct struct {
	namespace        string
	remoteCache      bool
	asyncRemoteCache bool
	lock             sync.RWMutex

	lruCache    *lru.Cache
	redisClient redis.UniversalClient

	stat       *cacheStat
	ctx        context.Context
	cancelFunc func()

	stream     chan cacheItem
	serializer codec.ICodec
}

// NewOneCacheStructWithNamespace --
func NewOneCacheStructWithNamespace(namespace string) *OneCacheStruct {
	lruCache, err := lru.NewWithExpiration(DefaultMaxEntries, DefaultTimeoutInSeconds)
	if err != nil {
		logger.Panicf("Cannot init lru cache: %v", err)
		return nil
	}

	ctx, cancelFunc := util.GetContextWithCancel(context.Background())

	return &OneCacheStruct{
		namespace:   namespace,
		remoteCache: false,
		lruCache:    lruCache,
		stat:        &cacheStat{},
		ctx:         ctx,
		cancelFunc:  cancelFunc,
		serializer:  &codec.JSONCodec{},
		stream:      make(chan cacheItem, 0),
	}
}

// NewOneCacheStruct --
func NewOneCacheStruct() *OneCacheStruct {
	lruCache, err := lru.NewWithExpiration(DefaultMaxEntries, DefaultTimeoutInSeconds)
	if err != nil {
		logger.Panicf("Cannot init lru cache: %v", err)
		return nil
	}

	ctx, cancelFunc := util.GetContextWithCancel(context.Background())

	return &OneCacheStruct{
		remoteCache: false,
		lruCache:    lruCache,
		stat:        &cacheStat{},
		ctx:         ctx,
		cancelFunc:  cancelFunc,
		serializer:  &codec.JSONCodec{},
		stream:      make(chan cacheItem, 0),
	}
}

// SetRedis --
func (o *OneCacheStruct) SetRedis(redisClient redis.UniversalClient) *OneCacheStruct {
	o.redisClient = redisClient
	o.remoteCache = true
	o.asyncRemoteCache = true

	logger.Infof("Enable remote synced")
	o.remoteSync()

	return o
}

// SetSize --
func (o *OneCacheStruct) SetSize(maxSize int) *OneCacheStruct {
	lruCache, _ := lru.NewWithExpiration(maxSize, DefaultTimeoutInSeconds)

	o.lruCache = lruCache
	return o
}

// SetContext --
func (o *OneCacheStruct) SetContext(ctx context.Context) *OneCacheStruct {
	ctx, cancelFunc := util.GetContextWithCancel(ctx)
	o.ctx = ctx
	o.cancelFunc = cancelFunc
	return o
}

// SetNamespace --
func (o *OneCacheStruct) SetNamespace(namespace string) *OneCacheStruct {
	o.namespace = namespace
	return o
}

// SetSerializer --
func (o *OneCacheStruct) SetSerializer(c codec.ICodec) *OneCacheStruct {
	o.serializer = c
	return o
}

// Init Redis
func (o *OneCacheStruct) set(key string, value interface{}, ttl time.Duration) {
	defer atomic.AddUint32(&o.stat.totalCount, 1)

	if ttl <= 0 {
		ttl = DefaultTimeoutInSeconds
	}

	valueStr, err := o.serializer.Encode(value)

	if err != nil {
		logger.Errorf("Cannot encode to byte: %v", err)
		return
	}

	o.lruCache.Add(key, valueStr, ttl)

	atomic.AddUint32(&o.stat.setCount, 1)

	if o.remoteCache && o.asyncRemoteCache {
		o.stream <- cacheItem{key, valueStr, ttl, KVType}
	}
}

func (o *OneCacheStruct) get(key string, dataType int) (*Item, error) {
	defer atomic.AddUint32(&o.stat.totalCount, 1)

	start := time.Now()
	value, isExisted := o.lruCache.Get(key)

	if duration := time.Since(start); duration > OptimalInMemAccessTime {
		logger.Warnf("%v cache has reach optimal access time for key %v: %v", o.namespace, key, duration)
	}

	if !isExisted {
		if o.asyncRemoteCache {
			remoteKey := fmt.Sprintf("%v_%v", o.namespace, key)
			val, err := o.redisClient.Get(remoteKey).Result()

			if duration := time.Since(start); duration > OptimalInMemAccessTime {
				logger.Warnf("%v cache has reach optimal access time for key %v: %v", o.namespace, key, duration)
			}

			if err != nil {
				if err.Error() == redis.Nil.Error() {
					atomic.AddUint32(&o.stat.missCount, 1)
				}
				return nil, err
			}

			atomic.AddUint32(&o.stat.hitCount, 1)

			valItem, _ := o.serializer.Decode([]byte(val))
			return &Item{valItem, []byte(val)}, nil
		}

		atomic.AddUint32(&o.stat.missCount, 1)
		return nil, ErrNotFound
	}

	atomic.AddUint32(&o.stat.hitCount, 1)
	valItem, _ := o.serializer.Decode(value.([]byte))
	return &Item{valItem, value.([]byte)}, nil
}

func (o *OneCacheStruct) remoteSync() {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case item := <-o.stream:
				remoteKey := fmt.Sprintf("%v_%v", o.namespace, item.Key)
				var err error

				switch item.DataType {
				case KVType:
					err = o.redisClient.Set(remoteKey, item.Value, item.TTL).Err()
				case SetType:
				case SortedSetType:

				}

				if err != nil && err.Error() != redis.Nil.Error() {
					logger.Errorf("Cannot fetch data from redis: %v", err)
				}

			}
		}
	}(o.ctx)
}

// Set --
func (o *OneCacheStruct) Set(key string, value interface{}, ttlInSeconds int) {
	if ttlInSeconds < 0 {
		ttlInSeconds = -1
	}

	o.set(key, value, time.Second*time.Duration(ttlInSeconds))
}

// Get --
func (o *OneCacheStruct) Get(key string) (*Item, error) {
	val, err := o.get(key, KVType)
	return val, err
}
