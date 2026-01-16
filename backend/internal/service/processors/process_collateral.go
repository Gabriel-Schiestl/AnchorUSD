package processors

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

var collateralKeysToChange = []string{
	"collateral:total_supply",
	"user:collateral",
}

func ProcessCollateral(metric model.Metrics, cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	amountToChange := getAmountChange(metric)

	collateralKeysWithUserAddress := make([]string, len(collateralKeysToChange))
	copy(collateralKeysWithUserAddress, collateralKeysToChange)

	collateralKeysWithUserAddress[1] = collateralKeysWithUserAddress[1] + ":" + metric.UserAddress.Hex() + ":" + metric.CollateralTokenAddress.Hex()

	cacheStore.MultiAdd(collateralKeysWithUserAddress, amountToChange)

	getCollateralUSDAmount, err := getCollateralUSDAmount(metric, priceFeed, cacheStore)
	if err != nil {
		return
	}

	debt, err := cacheStore.Get("user:debt:" + metric.UserAddress.Hex())
	if err != nil {
		return
	}

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	healthFactor := domain.CalculateHealthFactor(getCollateralUSDAmount, debtBigInt)

	cacheStore.Set("user:collateral_usd:"+metric.UserAddress.Hex(), getCollateralUSDAmount.String(), 0)
	cacheStore.Set("user:health_factor:"+metric.UserAddress.Hex(), healthFactor.String(), 0)
}

func getAmountChange(metric model.Metrics) *big.Int {
	amountToChange := metric.Amount
	if metric.Operation == model.Subtraction {
		amountToChange = big.NewInt(0).Neg(amountToChange) 
	}
	return amountToChange
}

func getCollateralUSDAmount(metric model.Metrics, priceFeed external.IPriceFeedAPI, cacheStore storage.ICacheStore) (*big.Int, error) {
	var totalUSDValue = big.NewInt(0)
	for name, token := range constants.CollateralTokens {
		price, err := getPrice(priceFeed, name)
		if err != nil {
			return nil, err
		}

		tokenAmount, err := cacheStore.Get("user:collateral:" + metric.UserAddress.Hex() + ":" + token)
		if err != nil {
			return nil, err
		}

		tokenAmountBigInt := big.NewInt(0)
		tokenAmountBigInt.SetString(tokenAmount, 10)

		tokenAmountInUSD, err := domain.GetTokenAmountInUSD(tokenAmountBigInt, price)
		if err != nil {
			return nil, err
		}
		totalUSDValue = big.NewInt(0).Add(totalUSDValue, tokenAmountInUSD)
	}

	return totalUSDValue, nil
}

func getPrice(priceFeed external.IPriceFeedAPI, name string) (string, error) {
	switch name {
	case "ETH":
		return priceFeed.GetEthUsdPrice()
	case "BTC":
		return priceFeed.GetBtcUsdPrice()
	default:
		return "0", nil
	}
}