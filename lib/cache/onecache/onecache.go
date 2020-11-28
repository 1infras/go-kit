package onecache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/1infras/go-kit/lib/cache/lru"
	"github.com/1infras/go-kit/lib/hook/common"
	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/util"
)

type OneCache interface {
	Set(ctx context.Context, key string, value interface{}, expiration int) error
	Get(ctx context.Context, key string) (Element, error)
	Contains(ctx context.Context, key string) bool
	Delete(ctx context.Context, key string)
	Flush(ctx context.Context) (err error)
	Report(ctx context.Context) (result string)
	AddHook(hook common.HookProcess)
}

const (
	DefaultNameSpace       = "one_cache"
	DefaultRemoteCache     = false
	DefaultMaxItems        = 100000
	DefaultTTL             = 30 * time.Second
	OptimalInMemAccessTime = 1 * time.Millisecond
)

var (
	Nil = errors.New("cache: key is not found")
)

type ClientOptionFunc func(*Client) error

// Caching with multiple layers include LRU and Redis
type Client struct {
	namespace   string
	remoteCache bool

	maxItems   int
	expiration time.Duration

	lru   lru.Client
	redis redis.UniversalClient

	stat       *stat
	context    context.Context
	cancelFunc func()

	stream chan item

	serializer Serializer

	lock sync.RWMutex
	hook *common.Hook
}

// NewCache
func NewOneCache(options ...ClientOptionFunc) (OneCache, error) {
	return NewClientWithContext(context.Background(), options...)
}

func NewClientWithContext(ctx context.Context, options ...ClientOptionFunc) (OneCache, error) {
	ctx, cancel := util.GetContextWithCancel(ctx)
	c := &Client{
		namespace:   DefaultNameSpace,
		remoteCache: DefaultRemoteCache,

		maxItems:   DefaultMaxItems,
		expiration: DefaultTTL,

		context:    ctx,
		cancelFunc: cancel,
		serializer: &DefaultSerializer{},

		stat: &stat{
			timeStart: time.Now().Unix(),
		},

		stream: make(chan item),
		hook: &common.Hook{},
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	lruCache, err := lru.NewWithExpiration(c.maxItems, c.expiration)
	if err != nil {
		return nil, fmt.Errorf("onecache: error with init lru cache: %v", err)
	}

	c.lru = lruCache

	if c.remoteCache {
		c.autoSyncRemoteCache()
	}

	return c, nil
}

func SetContext(ctx context.Context) ClientOptionFunc {
	return func(c *Client) error {
		if ctx == nil {
			return fmt.Errorf("context must not be empty")
		}
		ctx, cancel := util.GetContextWithCancel(ctx)
		c.context = ctx
		c.cancelFunc = cancel
		return nil
	}
}

func SetSerializer(serializer Serializer) ClientOptionFunc {
	return func(c *Client) error {
		if serializer == nil {
			return fmt.Errorf("serializer must not empty")
		}
		c.serializer = serializer
		return nil
	}
}

func SetRemoteCache(redisClient redis.UniversalClient) ClientOptionFunc {
	return func(c *Client) error {
		if redisClient == nil {
			return fmt.Errorf("redis client must be be empty")
		}

		_, err := redisClient.Ping(c.context).Result()
		if err != nil {
			return fmt.Errorf("connected to redis client has error: %v", err)
		}

		c.redis = redisClient
		c.remoteCache = true
		return nil
	}
}

func SetRemoteCacheNamespace(namespace string) ClientOptionFunc {
	return func(c *Client) error {
		if namespace == "" {
			return fmt.Errorf("namespace must not be empty")
		}
		c.namespace = namespace
		return nil
	}
}

func SetMaxItems(maxItems int) ClientOptionFunc {
	return func(c *Client) error {
		if maxItems <= 0 {
			return fmt.Errorf("max items must be greater than 0")
		}
		c.maxItems = maxItems
		return nil
	}
}

func SetExpiration(expiration time.Duration) ClientOptionFunc {
	return func(c *Client) error {
		if expiration <= 0 {
			return fmt.Errorf("expiration must be greater than 0")
		}
		c.expiration = expiration
		return nil
	}
}

func (_this *Client) AddHook(hook common.HookProcess) {
	_this.hook.AddHook(hook)
}

func (_this *Client) getRemoteKey(key string) string {
	return fmt.Sprintf("%s_%s", _this.namespace, key)
}

func (_this *Client) set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	defer func() {
		atomic.AddUint32(&_this.stat.totalWrites, 1)
		atomic.AddUint32(&_this.stat.totalOperations, 1)
	}()

	if expiration < 0 {
		expiration = _this.expiration
	}

	b, err := _this.serializer.Encode(value)

	if err != nil {
		return fmt.Errorf("encode value has error: %v", err)
	}

	_this.lru.Add(ctx, key, b, expiration)

	if _this.remoteCache {
		_this.stream <- item{
			key:        key,
			value:      b,
			expiration: expiration,
			action:     AddElement,
		}
	}

	atomic.AddInt64(&_this.stat.totalWriteBytes, int64(len(b)))

	return nil
}

func (_this *Client) get(ctx context.Context, key string) (Element, error) {
	defer func() {
		atomic.AddUint32(&_this.stat.totalReads, 1)
		atomic.AddUint32(&_this.stat.totalOperations, 1)
	}()

	var (
		rawValue    interface{}
		encodeValue []byte
		existed     bool
		err         error
	)

	start := time.Now()
	rawValue, existed = _this.lru.Get(ctx, key)

	if duration := time.Since(start); duration > OptimalInMemAccessTime {
		logger.Warn("get cache has reach optimal access time", zap.String("namespace", _this.namespace), zap.String("key", key), zap.String("duration", duration.String()))
	}

	if !existed {
		if _this.remoteCache {
			rawValue, err = _this.redis.Get(ctx, _this.getRemoteKey(key)).Result()
			if duration := time.Since(start); duration > OptimalInMemAccessTime {
				logger.Warn("get cache has reach optimal access time", zap.String("namespace", _this.namespace), zap.String("key", key), zap.String("duration", duration.String()))
			}

			if err != nil {
				if err.Error() == redis.Nil.Error() {
					atomic.AddUint32(&_this.stat.totalMisses, 1)
				}
				return nil, Nil
			}

			atomic.AddUint32(&_this.stat.totalHits, 1)

			encodeValue, err = _this.serializer.Encode(rawValue)
			if err != nil {
				return nil, fmt.Errorf("encode has error: %v", err)
			}
			goto Return
		}

		atomic.AddUint32(&_this.stat.totalMisses, 1)
		return nil, Nil
	} else {
		encodeValue = rawValue.([]byte)
	}

	atomic.AddUint32(&_this.stat.totalHits, 1)

Return:
	atomic.AddInt64(&_this.stat.totalReadBytes, int64(len(encodeValue)))
	return &element{encodeValue}, nil
}

func (_this *Client) delete(ctx context.Context, key string) {
	defer atomic.AddUint32(&_this.stat.totalOperations, 1)

	_this.lru.Remove(ctx, key)

	if _this.remoteCache {
		_this.stream <- item{
			key:    key,
			action: DeleteElement,
		}
	}
}

func (_this *Client) contains(ctx context.Context, key string) bool {
	defer atomic.AddUint32(&_this.stat.totalOperations, 1)

	if ok := _this.lru.Contains(ctx, key); ok {
		return ok
	}

	if _this.remoteCache {
		ok, err := _this.redis.Exists(ctx, key).Result()
		if err != nil && err.Error() != redis.Nil.Error() {
			logger.Warn("get exist remote cache with redis has error", zap.String("key", key), zap.String("error", err.Error()))
			return false
		}
		return ok == 1
	}

	return false
}

func (_this *Client) flush(ctx context.Context) error {
	_this.lru.Purge(ctx)

	if _this.remoteCache {
		var (
			cursor     uint64
			remoteKeys []string
			err        error
		)

		for {
			var keys []string
			keys, cursor, err = _this.redis.Scan(ctx, cursor, _this.namespace, 10).Result()
			if err != nil {
				return err
			}

			if cursor == 0 {
				break
			}

			remoteKeys = append(remoteKeys, keys...)
		}

		if len(remoteKeys) > 0 {
			_, err = _this.redis.Del(ctx, remoteKeys...).Result()
			if err != nil {
				return err
			}
		}
	}

	_this.stat.reset()

	return nil
}

func (_this *Client) autoSyncRemoteCache() {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case item := <-_this.stream:
				key := _this.getRemoteKey(item.key)
				switch item.action {
				case AddElement:
					err := _this.redis.Set(_this.context, key, item.value, item.expiration).Err()
					if err != nil && err.Error() != redis.Nil.Error() {
						logger.Error("set remote cache with redis has error", zap.String("key", key), zap.String("error", err.Error()))
					}
				case DeleteElement:
					_, err := _this.redis.Del(_this.context, key).Result()
					if err != nil && err.Error() != redis.Nil.Error() {
						logger.Error("delete remote cache with redis has error", zap.String("key", key), zap.String("error", err.Error()))
					}
				}

			}
		}
	}(_this.context)
}

func (_this *Client) report() string {
	duration := time.Since(time.Unix(_this.stat.timeStart, 0)).Seconds()

	return fmt.Sprintf(`
		#Cache Stat
		Total ops: %v,
		Total write ops: %v,
		Total read ops: %v,
		Total other ops: %v,
		Total write bytes: %v,
		Total read bytes: %v,
		Read ops per second: %.2f,
		Write ops per second: %.2f,
		Total hits: %v,
		Total misses: %v`,
		_this.stat.totalOperations,
		_this.stat.totalWrites,
		_this.stat.totalReads,
		_this.stat.totalOperations-_this.stat.totalWrites-_this.stat.totalReads,
		_this.stat.totalWriteBytes,
		_this.stat.totalReadBytes,
		float64(_this.stat.totalWrites*1.0)/duration,
		float64(_this.stat.totalReads*1.0)/duration,
		_this.stat.totalHits,
		_this.stat.totalMisses)
}

func (_this *Client) Set(ctx context.Context, key string, value interface{}, expiration int) (err error) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	if expiration < 0 {
		expiration = -1
	}

	_this.hook.Process(ctx, func() {
		err = _this.set(ctx, key, value, time.Second*time.Duration(expiration))
	}, "set")

	return
}

func (_this *Client) Get(ctx context.Context, key string) (element Element, err error) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		element, err = _this.get(ctx, key)
	}, "get")

	return
}

func (_this *Client) Delete(ctx context.Context, key string) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		_this.delete(ctx, key)
	}, "delete")
}

func (_this *Client) Contains(ctx context.Context, key string) (existed bool) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		existed = _this.contains(ctx, key)
	}, "contains")

	return
}

func (_this *Client) Flush(ctx context.Context) (err error) {
	_this.lock.Lock()
	defer _this.lock.Unlock()

	_this.hook.Process(ctx, func() {
		err = _this.flush(ctx)
	}, "flush")

	return
}

func (_this *Client) Report(ctx context.Context) (result string) {
	_this.hook.Process(ctx, func() {
		result = _this.report()
	}, "report")

	return
}
