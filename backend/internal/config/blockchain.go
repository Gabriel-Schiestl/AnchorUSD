package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type blockchainConfig struct {
	ProviderURL string `env:"BLOCKCHAIN_PROVIDER_URL"`
}

func GetBlockchainConfig() *blockchainConfig {
	cfg := &blockchainConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("Failed to parse blockchain config: %v", err)
	}
	return cfg
}

func (bc *blockchainConfig) GetProviderURL() string {
	return bc.ProviderURL
}