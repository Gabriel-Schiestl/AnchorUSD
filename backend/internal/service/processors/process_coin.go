package processors

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func ProcessCoin(metric model.Metrics, cacheStore storage.ICacheStore) {
	amountToChange := getAmountChange(metric)

	cacheStore.HAdd("coin", "total_supply", amountToChange)

	cacheStore.HAdd("user:debt", metric.UserAddress.Hex(), amountToChange)

	collateralUSDValue, err := cacheStore.HGet("user:collateral_usd", metric.UserAddress.Hex())
	if err != nil {
		return
	}

	debt, err := cacheStore.HGet("user:debt", metric.UserAddress.Hex())
	if err != nil {
		return
	}

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	collateralUSDBigInt := big.NewInt(0)
	collateralUSDBigInt.SetString(collateralUSDValue, 10)

	healthFactor := domain.CalculateHealthFactor(collateralUSDBigInt, debtBigInt)

	cacheStore.HSet("user:health_factor", metric.UserAddress.Hex(), healthFactor.String())
}
