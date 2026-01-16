package model

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Asset string

type Operation string

const (
	CollateralAsset Asset = "collateral"
	StablecoinAsset Asset = "stablecoin"
	Addition        Operation = "addition"
	Subtraction     Operation = "subtraction"
)

type Metrics struct {
	UserAddress common.Address
	Amount      *big.Int
	Asset       Asset
	Operation   Operation
}
