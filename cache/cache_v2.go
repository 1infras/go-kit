package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	// DefaultTTL --
	DefaultTTL = 30 * time.Second

	// OptimalInMemAccessTime // Follow Jeff Dean, read 1MB sequential from memory ideally is 250 micro-seconds --
	OptimalInMemAccessTime = 1 * time.Millisecond

	// OptimalNetworkAccessTime --
	OptimalNetworkAccessTime = 150 * time.Millisecond

	// ErrNotFound --
	ErrNotFound = errors.New("cache: Key is not found")

	// ReportInSecond --
	ReportInSecond = 60 * time.Second
)

type cacheStat struct {
	hitCount   uint32
	missCount  uint32
	totalCount uint32
	// expiredCount uint32
	setCount uint32
	// totalSize    uint32

	totalReadsProcessed  uint32
	totalWritesProcessed uint32

	startTime int64

	totalReadBytes  int64
	totalWriteBytes int64
}

type cacheItem struct {
	Key      string
	Value    interface{}
	TTL      time.Duration
	DataType int
}

// Item --
type Item struct {
	value []byte
}

// Bytes --
func (i *Item) Bytes() []byte {
	return i.value
}

// Val --
func (i *Item) Val() interface{} {
	return i.value
}

// Int64 --
func (i *Item) Int64() int64 {
	v, _ := strconv.ParseInt(string(i.value), 10, 64)
	return v
}

// Int --
func (i *Item) Int() int {
	v, _ := strconv.Atoi(string(i.value))
	return v
}

// Boolean --
func (i *Item) Boolean() bool {
	v, _ := strconv.ParseBool(string(i.value))
	return v
}

// Float64 --
func (i *Item) Float64() float64 {
	v, _ := strconv.ParseFloat(string(i.value), 64)
	return v
}

// String --
func (i *Item) String() string {
	return string(i.value)
}

// OneCacheStruct --
type OneCacheStruct struct {
	namespace        string
	remoteCache      bool
	asyncRemoteCache bool

	lruCache    *lru.Cache
	redisClient redis.UniversalClient

	stat       *cacheStat
	ctx        context.Context
	cancelFunc func()

	stream     chan cacheItem
	serializer codec.ICodec
}

// NewOneCacheStruct --
func NewOneCacheStruct() *OneCacheStruct {
	lruCache, err := lru.NewWithExpiration(DefaultMaxEntries, DefaultTTL)
	if err != nil {
		logger.Panicf("Cannot init lru cache: %v", err)
		return nil
	}

	ctx, cancelFunc := util.GetContextWithCancel(context.Background())

	return &OneCacheStruct{
		remoteCache: false,
		lruCache:    lruCache,
		ctx:         ctx,
		cancelFunc:  cancelFunc,
		serializer:  &codec.JSONCodec{},
		stream:      make(chan cacheItem),
		stat: &cacheStat{
			startTime: time.Now().Unix(),
		},
	}
}

// RunReport --
func (o *OneCacheStruct) RunReport() {
	ticker := time.NewTicker(ReportInSecond)
	go func() {
		for {
			select {
			case <-o.ctx.Done():
				return
			case <-ticker.C:
				fmt.Println(o.Report())
			}
		}
	}()
}

// WithRedis --
func (o *OneCacheStruct) WithRedis(redisClient redis.UniversalClient, namespace string) *OneCacheStruct {
	o.redisClient = redisClient
	o.remoteCache = true
	o.asyncRemoteCache = true
	o.namespace = namespace

	logger.Infof("Enable remote synced")
	o.remoteSync()

	return o
}

// WithContext --
func (o *OneCacheStruct) WithContext(ctx context.Context) *OneCacheStruct {
	ctx, cancelFunc := util.GetContextWithCancel(ctx)
	o.ctx = ctx
	o.cancelFunc = cancelFunc
	return o
}

// SetSerializer --
func (o *OneCacheStruct) SetSerializer(c codec.ICodec) *OneCacheStruct {
	o.serializer = c
	return o
}

// Init Redis
func (o *OneCacheStruct) set(key string, value interface{}, ttl time.Duration) {
	defer atomic.AddUint32(&o.stat.totalWritesProcessed, 1)

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	valueBytes, err := o.serializer.Encode(value)

	if err != nil {
		logger.Errorf("Cannot encode to byte: %v", err)
		return
	}

	o.lruCache.Add(key, valueBytes, ttl)

	if o.remoteCache && o.asyncRemoteCache {
		o.stream <- cacheItem{key, valueBytes, ttl, KVType}
	}

	atomic.AddUint32(&o.stat.setCount, 1)
	atomic.AddInt64(&o.stat.totalWriteBytes, int64(len(valueBytes)))
}

func (o *OneCacheStruct) get(key string, dataType int) (*Item, error) {
	defer atomic.AddUint32(&o.stat.totalReadsProcessed, 1)

	var valueBytes []byte

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

			valueBytes = []byte(val)
			goto Return
		}

		atomic.AddUint32(&o.stat.missCount, 1)
		return nil, ErrNotFound
	} else {
		valueBytes = value.([]byte)
	}

	atomic.AddUint32(&o.stat.hitCount, 1)

Return:

	atomic.AddInt64(&o.stat.totalWriteBytes, int64(len(valueBytes)))
	return &Item{valueBytes}, nil
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

// Delete --
func (o *OneCacheStruct) Delete(key string) error {
	o.lruCache.Remove(key)

	if o.asyncRemoteCache {
		_, err := o.redisClient.Del(fmt.Sprintf("%v_%v", o.namespace, key)).Result()
		return err
	}

	return nil
}

// Flush --
func (o *OneCacheStruct) Flush() error {
	o.lruCache.Purge()

	if o.remoteCache {
		var cursor uint64
		var allKeys []string
		var err error

		for {
			var keys []string
			keys, cursor, err = o.redisClient.Scan(cursor, "*", 10).Result()
			if err != nil {
				return err
			}

			if cursor == 0 {
				break
			}
			allKeys = append(allKeys, keys...)

		}
		_, err = o.redisClient.Del(allKeys...).Result()
		return err
	}

	return nil
}

// Exists --
func (o *OneCacheStruct) Exists(key string) bool {
	isExistedLocal := o.lruCache.Contains(key)
	if !isExistedLocal && o.asyncRemoteCache {
		isRemote, err := o.redisClient.Exists(fmt.Sprintf("%v_%v", o.namespace, key)).Result()
		if err != nil && err.Error() != redis.Nil.Error() {
			logger.Errorf("Cannot fetch data from redis: %v", err)
			return false
		}
		return isRemote == 1
	}

	return isExistedLocal
}

// Report --
func (o *OneCacheStruct) Report() string {
	duration := time.Since(time.Unix(o.stat.startTime, 0)).Seconds()

	return fmt.Sprintf(`
		#Report
		Total ops: %v,
		Total write ops: %v,
		Total read ops: %v,
		Total write bytes: %v,
		Total read bytes: %v,
		Read ops per second: %.2f,
		Write ops per second: %.2f,
		Cache hits: %v,
		Cache misses: %v
	`, o.stat.totalCount,
		o.stat.totalWritesProcessed,
		o.stat.totalReadsProcessed,
		o.stat.totalWriteBytes,
		o.stat.totalReadBytes,
		float64(o.stat.totalWritesProcessed*1.0)/duration,
		float64(o.stat.totalReadsProcessed*1.0)/duration,
		o.stat.hitCount,
		o.stat.missCount,
	)
}

// UnmarshalKey --
func (o *OneCacheStruct) UnmarshalKey(key string, obj interface{}) error {
	val, err := o.get(key, KVType)
	if err != nil {
		return err
	}

	return o.serializer.Decode(val.Bytes(), obj)
}
