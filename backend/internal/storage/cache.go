package storage

import (
	"math/big"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var multiAddScript = redis.NewScript(`
	for i = 1, #KEYS do
		redis.call("INCRBY", KEYS[i], ARGV[1])
	end
`)

type CacheConfig interface {
	GetAddress() string
	GetPassword() string
}

type CacheStore struct {
	Client *redis.Client
	mu sync.RWMutex
}

type ICacheStore interface {
	Get(key string) (string, error)
	Set(key string, value any, expiration time.Duration) (string, error)
	Increment(key string, amountInWei *big.Int) (int64 , error)
	Decrement(key string, amountInWei *big.Int) (int64 , error)
	MultiAdd(keys []string, amountInWei *big.Int) error
}

func NewCacheStore(config CacheConfig) *CacheStore {
	return &CacheStore{
		Client: redis.NewClient(&redis.Options{
			Addr: config.GetAddress(),
			Password: config.GetPassword(),
		}),
	}
}

func (cs *CacheStore) Get(key string) (string, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.Client.Get(key).Result()
}

func (cs *CacheStore) Set(key string, value any, expiration time.Duration) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.Client.Set(key, value, expiration).Result()
}

func (cs *CacheStore) Increment(key string, amountInWei *big.Int) (int64 , error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.Client.IncrBy(key, amountInWei.Int64()).Result()
}

func (cs *CacheStore) Decrement(key string, amountInWei *big.Int) (int64 , error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.Client.DecrBy(key, amountInWei.Int64()).Result()
}

func (cs *CacheStore) MultiAdd(keys []string, amountInWei *big.Int) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if amountInWei.Sign() == 0 {
		return nil
	}

	_, err := multiAddScript.Run(
		cs.Client,
		keys,
		amountInWei.String(),
	).Result()

	return err
}