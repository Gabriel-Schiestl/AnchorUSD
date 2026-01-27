package domain

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func CalculateMaxMintable(collateralValueUSDStr, totalDebtStr string) string {
	collateralValueUSD := new(big.Int)
	collateralValueUSD.SetString(collateralValueUSDStr, 10)

	collateralValueUSD.Mul(collateralValueUSD, constants.ADDITIONAL_PRICE_PRECISION)

	totalDebt := new(big.Int)
	totalDebt.SetString(totalDebtStr, 10)

	collateralAdjusted := new(big.Int).Mul(collateralValueUSD, constants.LIQUIDATION_THRESHOLD)
	collateralAdjusted.Div(collateralAdjusted, constants.LIQUIDATION_PRECISION)

	maxMintable := new(big.Int).Sub(collateralAdjusted, totalDebt)

	if maxMintable.Sign() < 0 {
		return "0"
	}

	return maxMintable.String()
}