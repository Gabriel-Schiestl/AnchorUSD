package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
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
	Amount      decimal.Decimal
	Asset       Asset
	Operation   Operation
}
