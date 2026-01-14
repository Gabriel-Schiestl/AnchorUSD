package blockchain

import "github.com/ethereum/go-ethereum/ethclient"

type BlockchainProvider interface {
	GetProviderURL() string
}

func GetClient(provider BlockchainProvider) *ethclient.Client {
	client, err := ethclient.Dial(provider.GetProviderURL())
	if err != nil {
		panic(err)
	}

	return client
}