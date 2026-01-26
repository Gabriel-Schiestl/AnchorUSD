package processors

import (
	"fmt"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

func ProcessCollateral(metric model.Metrics, cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	logger := utils.GetLogger()
	logger.Info().Str("user", metric.UserAddress.Hex()).Str("token", metric.CollateralTokenAddress.Hex()).Str("amount", metric.Amount.String()).Str("operation", string(metric.Operation)).Msg("Processing collateral metric")

	amountToChange := getAmountChange(metric)
	logger.Debug().Str("amount_to_change", amountToChange.String()).Msg("Calculated amount to change")

	collateralKey := "collateral:" + metric.CollateralTokenAddress.Hex()

	cacheStore.HAdd(collateralKey, metric.UserAddress.Hex(), amountToChange)
	logger.Debug().Str("key", collateralKey).Str("user", metric.UserAddress.Hex()).Msg("Updated user collateral balance")

	getCollateralUSDAmount, err := getCollateralUSDAmount(metric, priceFeed, priceStore, cacheStore)
	if err != nil {

		fmt.Println("Error getting collateral USD amount:", err)
		logger.Error().Err(err).Str("user", metric.UserAddress.Hex()).Msg("Failed to get collateral USD amount")
		return
	}
	logger.Debug().Str("user", metric.UserAddress.Hex()).Str("collateral_usd", getCollateralUSDAmount.String()).Msg("Calculated total collateral USD value")

	debt, err := cacheStore.HGet("user:debt", metric.UserAddress.Hex())
	if err != nil {
		if debt != "" {
			logger.Error().Err(err).Str("user", metric.UserAddress.Hex()).Msg("Failed to get user debt")
			return
		}
		debt = "0"
	}

	usdAmountToChange, err := getUSDAmountToChange(metric, priceFeed, priceStore)
	if err != nil {
		logger.Error().Err(err).Str("user", metric.UserAddress.Hex()).Msg("Failed to get USD amount to change")
		return
	}
	logger.Debug().Str("usd_change", usdAmountToChange.String()).Msg("Calculated USD amount to change")

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	logger.Debug().Str("user", metric.UserAddress.Hex()).Str("collateral_usd", getCollateralUSDAmount.String()).Str("debt", debt).Msg("Calculating health factor")

	healthFactor := domain.CalculateHealthFactor(getCollateralUSDAmount, debtBigInt)

	cacheStore.HAdd("collateral", "total_supply", usdAmountToChange)
	cacheStore.HSet("user:collateral_usd", metric.UserAddress.Hex(), getCollateralUSDAmount.String())
	cacheStore.HSet("user:health_factor", metric.UserAddress.Hex(), healthFactor.String())

	logger.Info().Str("user", metric.UserAddress.Hex()).Str("health_factor", healthFactor.String()).Str("collateral_usd", getCollateralUSDAmount.String()).Msg("Collateral metric processed and health factor updated")
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
			if tokenAmount != "" {
				return nil, err
			}
			tokenAmount = "0"
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
