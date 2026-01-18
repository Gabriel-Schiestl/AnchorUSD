package processors

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

var coinKeysToChange = []string{
	"coin:total_supply",
	"user:debt",
}

func ProcessCoin(metric model.Metrics, cacheStore storage.ICacheStore) {
	amountToChange := getAmountChange(metric)

	coinKeysWithUserAddress := make([]string, len(coinKeysToChange))
	copy(coinKeysWithUserAddress, coinKeysToChange)

	coinKeysWithUserAddress[1] = coinKeysWithUserAddress[1] + ":" + metric.UserAddress.Hex()

	cacheStore.MultiAdd(coinKeysWithUserAddress, amountToChange)

	collateralUSDValue, err := cacheStore.Get("user:collateral_usd:" + metric.UserAddress.Hex())
	if err != nil {
		return
	}

	debt, err := cacheStore.Get("user:debt:" + metric.UserAddress.Hex())
	if err != nil {
		return
	}

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	collateralUSDBigInt := big.NewInt(0)
	collateralUSDBigInt.SetString(collateralUSDValue, 10)

	healthFactor := domain.CalculateHealthFactor(collateralUSDBigInt, debtBigInt)

	cacheStore.Set("user:health_factor:"+metric.UserAddress.Hex(), healthFactor.String(), 0)
}
