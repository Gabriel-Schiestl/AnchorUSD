package domain

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func CalculateLiquidationAmount(debt *big.Int) *big.Int {
	if debt == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Div(new(big.Int).Set(debt), constants.LIQUIDATION_DIVISOR)
}

func PercentageOf(value, total *big.Int) float64 {
	if total == nil || total.Sign() == 0 {
		return 0.0
	}
	pctBig := new(big.Int).Mul(value, constants.PERCENTAGE_MULTIPLIER)
	pctBig.Div(pctBig, total)

	return float64(pctBig.Int64()) / constants.PERCENTAGE_BASE_DIVISOR
}



