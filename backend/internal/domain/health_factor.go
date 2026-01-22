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

func CalculateHealthFactorAfterMint(currentCollateralUSD, currentDebt, mintAmount *big.Int) *big.Int {
	newDebt := new(big.Int).Add(currentDebt, mintAmount)
	return CalculateHealthFactor(currentCollateralUSD, newDebt)
}

func CalculateHealthFactorAfterBurn(currentCollateralUSD, currentDebt, burnAmount *big.Int) *big.Int {
	newDebt := new(big.Int).Sub(currentDebt, burnAmount)
	if newDebt.Sign() < 0 {
		newDebt = big.NewInt(0)
	}
	return CalculateHealthFactor(currentCollateralUSD, newDebt)
}

func CalculateHealthFactorAfterDeposit(currentCollateralUSD, currentDebt, depositAmountUSD *big.Int) *big.Int {
	newCollateralUSD := new(big.Int).Add(currentCollateralUSD, depositAmountUSD)
	return CalculateHealthFactor(newCollateralUSD, currentDebt)
}

func AverageHealthFactor(sumHF *big.Int, totalUsers int) float64 {
	if totalUsers == 0 {
		return 0.0
	}
	avg := new(big.Int).Div(new(big.Int).Set(sumHF), big.NewInt(int64(totalUsers)))
	return float64(avg.Int64()) / float64(constants.PRECISION.Int64())
}

func IsAtRisk(hf *big.Int) bool {
	if hf == nil {
		return false
	}
	return hf.Cmp(constants.RISK_THRESHOLD) < 0
}