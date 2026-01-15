package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type cacheConfig struct {
	Address  string `env:"CACHE_ADDRESS"`
	Password string `env:"CACHE_PASSWORD"`
}

func GetCacheConfig() *cacheConfig {
	cfg := &cacheConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Failed to parse cache config: %v", err)
	}
	return cfg
}

func (bc *cacheConfig) GetAddress() string {
	return bc.Address
}

func (bc *cacheConfig) GetPassword() string {
	return bc.Password
}