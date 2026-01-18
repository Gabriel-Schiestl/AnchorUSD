package domain

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func CalculateHealthFactor(collateralValueUSD, debtValueUSD *big.Int) *big.Int {
	if debtValueUSD.Sign() == 0 {
		return big.NewInt(0).Set(constants.PRECISION)
	}

	collateralAdjusted := big.NewInt(0).Mul(collateralValueUSD, constants.LIQUIDATION_THRESHOLD)
	collateralAdjusted.Div(collateralAdjusted, constants.LIQUIDATION_PRECISION)
	healthFactor := big.NewInt(0).Mul(collateralAdjusted, constants.PRECISION)
	healthFactor.Div(healthFactor, debtValueUSD)

	return healthFactor
}
