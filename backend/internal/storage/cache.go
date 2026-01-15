package storage

import (
	"time"

	"github.com/go-redis/redis"
)

type CacheConfig interface {
	GetAddress() string
	GetPassword() string
}

type CacheStore struct {
	Client *redis.Client
}

func NewCacheStore(config CacheConfig) *CacheStore {
	return &CacheStore{
		Client: redis.NewClient(&redis.Options{
			Addr: config.GetAddress(),
			Password: config.GetPassword(),
		}),
	}
}

func (cs *CacheStore) Get(key string) *redis.StringCmd {
	return cs.Client.Get(key)
}

func (cs *CacheStore) Set(key string, value any, expiration time.Duration) *redis.StatusCmd {
	return cs.Client.Set(key, value, expiration)
}