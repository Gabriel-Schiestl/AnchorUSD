package blockchain

import (
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var contractABI *abi.ABI

func loadABI() (*abi.ABI, error) {
	f, err := os.Open("../../AUSDEngine.abi.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parsedABI, err := abi.JSON(f)
	if err != nil {
		return nil, err
	}

	return &parsedABI, nil
}

func GetABI() (*abi.ABI, error) {
	if contractABI != nil {
		return contractABI, nil
	}

	contractABI, err := loadABI()
	if err != nil {
		return nil, err
	}

	return contractABI, nil
}
