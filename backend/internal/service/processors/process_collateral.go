package processors

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func ProcessCollateral(metric model.Metrics, cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	amountToChange := getAmountChange(metric)

	collateralKey := "collateral:" + metric.CollateralTokenAddress.Hex()

	cacheStore.HAdd(collateralKey, metric.UserAddress.Hex(), amountToChange)

	getCollateralUSDAmount, err := getCollateralUSDAmount(metric, priceFeed, priceStore, cacheStore)
	if err != nil {
		return
	}

	debt, err := cacheStore.HGet("user:debt", metric.UserAddress.Hex())
	if err != nil {
		return
	}

	usdAmountToChange, err := getUSDAmountToChange(metric, priceFeed, priceStore)
	if err != nil {
		return
	}

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	healthFactor := domain.CalculateHealthFactor(getCollateralUSDAmount, debtBigInt)

	cacheStore.HAdd("collateral", "total_supply", usdAmountToChange)
	cacheStore.HSet("user:collateral_usd", metric.UserAddress.Hex(), getCollateralUSDAmount.String())
	cacheStore.HSet("user:health_factor", metric.UserAddress.Hex(), healthFactor.String())
}

func getUSDAmountToChange(metric model.Metrics, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) (*big.Int, error) {
	var name string
	for tokenName, tokenAddress := range constants.CollateralTokens {
		if tokenAddress == metric.CollateralTokenAddress.Hex() {
			name = tokenName
			break
		}
	}

	price, err := getPrice(priceFeed, name, metric.BlockNumber, priceStore)
	if err != nil {
		return nil, err
	}

	usdAmount, err := domain.GetTokenAmountInUSD(metric.Amount, price)
	if err != nil {
		return nil, err
	}

	if metric.Operation == model.Subtraction {
		usdAmount = big.NewInt(0).Neg(usdAmount)
	}

	return usdAmount, nil
}

func getAmountChange(metric model.Metrics) *big.Int {
	amountToChange := metric.Amount
	if metric.Operation == model.Subtraction {
		amountToChange = big.NewInt(0).Neg(amountToChange)
	}
	return amountToChange
}

func getCollateralUSDAmount(metric model.Metrics, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore, cacheStore storage.ICacheStore) (*big.Int, error) {
	var totalUSDValue = big.NewInt(0)
	for name, token := range constants.CollateralTokens {
		price, err := getPrice(priceFeed, name, metric.BlockNumber, priceStore)
		if err != nil {
			return nil, err
		}

		collateralKey := "collateral:" + token
		tokenAmount, err := cacheStore.HGet(collateralKey, metric.UserAddress.Hex())
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

func getPrice(priceFeed external.IPriceFeedAPI, name string, blockNumber uint64, priceStore storage.IPriceStore) (string, error) {
	price, _ := priceStore.GetPriceInBlock(name, blockNumber)
	if price != nil {
		return *price, nil
	}

	result := "0"
	var err error

	switch name {
	case "ETH":
		result, err = priceFeed.GetEthUsdPrice()
	case "BTC":
		result, err = priceFeed.GetBtcUsdPrice()
	default:
	}

	errSave := priceStore.SavePriceInBlock(name, blockNumber, result)
	if errSave != nil {
		return "", errSave
	}

	return result, err
}
