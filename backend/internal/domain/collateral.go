package domain

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func CollateralizationRatio(totalCollateral, totalDebt *big.Int) float64 {
	if totalDebt == nil || totalDebt.Sign() == 0 {
		return 0.0
	}
	ratioBigInt := new(big.Int).Mul(totalCollateral, constants.PERCENTAGE_MULTIPLIER)
	ratioBigInt.Div(ratioBigInt, totalDebt)
	return float64(ratioBigInt.Int64()) / constants.PERCENTAGE_BASE_DIVISOR
}

func CalculateBackingPercentage(totalCollateral, circulating *big.Int) float64 {
	if circulating == nil || circulating.Sign() == 0 {
		return 0.0
	}
	backingBigInt := new(big.Int).Mul(totalCollateral, constants.PERCENTAGE_MULTIPLIER)
	backingBigInt.Div(backingBigInt, circulating)
	return float64(backingBigInt.Int64()) / constants.PERCENTAGE_BASE_DIVISOR
}