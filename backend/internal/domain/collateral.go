package domain

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
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

type CollateralAssetData struct {
	Name     string
	Amount   *big.Int
	PriceUSD string
}

func CalculateCollateralDeposited(assets []CollateralAssetData) []model.CollateralDeposited {
	collateralDeposited := []model.CollateralDeposited{}

	for _, asset := range assets {
		if asset.Amount == nil || asset.Amount.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		valueUsd, err := GetTokenAmountInUSD(asset.Amount, asset.PriceUSD)
		if err != nil {
			continue
		}

		collateralDeposited = append(collateralDeposited, model.CollateralDeposited{
			Asset:    asset.Name,
			Amount:   asset.Amount.String(),
			ValueUsd: valueUsd.String(),
		})
	}

	return collateralDeposited
}