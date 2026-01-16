package domain

import "math/big"

var LIQUIDATION_THRESHOLD = big.NewInt(50)
var LIQUIDATION_PRECISION = big.NewInt(100)
var PRECISION = big.NewInt(1e18)
var MIN_HEALTH_FACTOR = big.NewInt(1e18)
var MAX_HEALTH_FACTOR = big.NewInt(9e18)

func CalculateHealthFactor(collateralValueUSD, debtValueUSD *big.Int) *big.Int {
	if debtValueUSD.Sign() == 0 {
		return big.NewInt(0).Set(PRECISION) 
	}

	collateralAdjusted := big.NewInt(0).Mul(collateralValueUSD, LIQUIDATION_THRESHOLD)
	collateralAdjusted.Div(collateralAdjusted, LIQUIDATION_PRECISION)

	healthFactor := big.NewInt(0).Mul(collateralAdjusted, PRECISION)
	healthFactor.Div(healthFactor, debtValueUSD)

	return healthFactor
}