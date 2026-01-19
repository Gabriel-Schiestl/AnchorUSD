package storage

import (
	"errors"
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

var hIncrByBigIntScript = redis.NewScript(`
	return redis.call("HINCRBY", KEYS[1], ARGV[1], ARGV[2])
`)

var incrByBigIntScript = redis.NewScript(`
	return redis.call("INCRBY", KEYS[1], ARGV[1])
`)

type CacheConfig interface {
	GetAddress() string
	GetPassword() string
}

type CacheStore struct {
	Client *redis.Client
	mu     sync.RWMutex
}

type ICacheStore interface {
	Get(key string) (string, error)
	Set(key string, value any, expiration time.Duration) (string, error)
	Add(key string, amountInWei *big.Int) (*big.Int, error)
	MultiAdd(keys []string, amountInWei *big.Int) error
	HSet(key string, field string, value any) error
	HGet(key string, field string) (string, error)
	HAdd(key string, field string, amountInWei *big.Int) (*big.Int, error)
}

func NewCacheStore(config CacheConfig) *CacheStore {
	return &CacheStore{
		Client: redis.NewClient(&redis.Options{
			Addr:     config.GetAddress(),
			Password: config.GetPassword(),
		}),
	}
}

func (cs *CacheStore) Get(key string) (string, error) {
	return cs.Client.Get(key).Result()
}

func (cs *CacheStore) Set(key string, value any, expiration time.Duration) (string, error) {
	return cs.Client.Set(key, value, expiration).Result()
}

func (cs *CacheStore) Add(
	key string,
	amountInWei *big.Int,
) (*big.Int, error) {

	if amountInWei.Sign() == 0 {
		return big.NewInt(0), nil
	}

	res, err := incrByBigIntScript.Run(
		cs.Client,
		[]string{key},
		amountInWei.String(),
	).Result()

	if err != nil {
		return nil, err
	}

	switch v := res.(type) {
	case int64:
		return big.NewInt(v), nil
	case string:
		n, ok := new(big.Int).SetString(v, 10)
		if !ok {
			return nil, errors.New("invalid bigint returned from redis")
		}
		return n, nil
	default:
		return nil, errors.New("unexpected redis return type")
	}
}

func (cs *CacheStore) MultiAdd(keys []string, amountInWei *big.Int) error {
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

func (cs *CacheStore) HSet(key string, field string, value any) error {
	return cs.Client.HSet(key, field, value).Err()
}

func (cs *CacheStore) HGet(key string, field string) (string, error) {
	return cs.Client.HGet(key, field).Result()
}

func (cs *CacheStore) HAdd(
	key string,
	field string,
	amountInWei *big.Int,
) (*big.Int, error) {

	res, err := hIncrByBigIntScript.Run(
		cs.Client,
		[]string{key},
		field,
		amountInWei.String(),
	).Result()

	if err != nil {
		return nil, err
	}

	switch v := res.(type) {
	case int64:
		return big.NewInt(v), nil
	case string:
		n, ok := new(big.Int).SetString(v, 10)
		if !ok {
			return nil, errors.New("invalid bigint returned from redis")
		}
		return n, nil
	default:
		return nil, errors.New("unexpected redis return type")
	}
}

func (cs *CacheStore) HGetAll(key string) (map[string]string, error) {
	return cs.Client.HGetAll(key).Result()
}
