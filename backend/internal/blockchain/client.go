package blockchain

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainProvider interface {
	GetProviderURL() string
}

func GetClient(provider BlockchainProvider) *ethclient.Client {
	logger := utils.GetLogger()
	providerURL := provider.GetProviderURL()

	logger.Info().Str("provider_url", providerURL).Msg("Connecting to blockchain provider")

	client, err := ethclient.Dial(providerURL)
	if err != nil {
		logger.Fatal().Err(err).Str("provider_url", providerURL).Msg("Failed to connect to blockchain provider")
		panic(err)
	}

	logger.Info().Str("provider_url", providerURL).Msg("Successfully connected to blockchain")
	return client
}
