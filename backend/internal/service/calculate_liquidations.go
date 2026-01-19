package service

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func CalculateLiquidations(priceFeed external.IPriceFeedAPI, cacheStore storage.ICacheStore) error {
	totalUSDCollateralByUser := make(map[string]*big.Int)

	btcPrice, err := priceFeed.GetBtcUsdPrice()
	if err != nil {
		return err
	}

	ethPrice, err := priceFeed.GetEthUsdPrice()
	if err != nil {
		return err
	}

	for name, address := range constants.CollateralTokens {
		switch name {
		case "BTC":
			collateralSupplyByUser, err := cacheStore.HGetAll("collateral:"+address)
			if err != nil {
				continue
			}

			updateBTCCollateralMetrics(collateralSupplyByUser, btcPrice, totalUSDCollateralByUser)
		case "ETH":
			collateralSupplyByUser, err := cacheStore.HGetAll("collateral:" + address)
			if err != nil {
				continue
			}

			updateETHCollateralMetrics(collateralSupplyByUser, ethPrice, totalUSDCollateralByUser)
		}
	}

	totalCollateralSupply := new(big.Int)
	updateLiquidationHealthFactors(totalCollateralSupply, cacheStore, totalUSDCollateralByUser)

	err = cacheStore.HSet("collateral", "total_supply", totalCollateralSupply.String())
	if err != nil {
		return err
	}

	return nil
}

func updateBTCCollateralMetrics(collateralSupplyByUser map[string]string, btcPrice string, totalUSDCollateralByUser map[string]*big.Int) {
	for userAddress, collateralAmountStr := range collateralSupplyByUser {
		collateralAmount := new(big.Int)
		_, ok := collateralAmount.SetString(collateralAmountStr, 10)
        if !ok {
            continue
        }

		usdValue, err := domain.GetTokenAmountInUSD(collateralAmount, btcPrice)
		if err != nil {
			continue
		}

		if _, exists := totalUSDCollateralByUser[userAddress]; !exists {
			totalUSDCollateralByUser[userAddress] = big.NewInt(0)
		}

		totalUSDCollateralByUser[userAddress].Add(totalUSDCollateralByUser[userAddress], usdValue)
	}
}

func updateETHCollateralMetrics(collateralSupplyByUser map[string]string, ethPrice string, totalUSDCollateralByUser map[string]*big.Int) {
	for userAddress, collateralAmountStr := range collateralSupplyByUser {
		collateralAmount := new(big.Int)
		_, ok := collateralAmount.SetString(collateralAmountStr, 10)
        if !ok {
            continue
        }

		usdValue, err := domain.GetTokenAmountInUSD(collateralAmount, ethPrice)
		if err != nil {
			continue
		}

		if _, exists := totalUSDCollateralByUser[userAddress]; !exists {
			totalUSDCollateralByUser[userAddress] = big.NewInt(0)
		}

		totalUSDCollateralByUser[userAddress].Add(totalUSDCollateralByUser[userAddress], usdValue)
	}
}

func updateLiquidationHealthFactors(totalCollateralSupply *big.Int, cacheStore storage.ICacheStore, totalUSDCollateralByUser map[string]*big.Int) {
	for userAddress, totalCollateralUSD := range totalUSDCollateralByUser {
		err := cacheStore.HSet("user:collateral_usd", userAddress, totalCollateralUSD.String())
		if err != nil {
			continue
		}

		totalCollateralSupply.Add(totalCollateralSupply, totalCollateralUSD)

		debt, err := cacheStore.HGet("user:debt", userAddress)
		if err != nil {
			continue
		}

		debtBigInt := new(big.Int)
		_, ok := debtBigInt.SetString(debt, 10)
        if !ok {
            continue
        }

		healthFactor := domain.CalculateHealthFactor(totalCollateralUSD, debtBigInt)
		if healthFactor.Cmp(constants.MIN_HEALTH_FACTOR) < 0 {
			err := cacheStore.HSet("liquidatable", userAddress, healthFactor.String())
			if err != nil {
				continue
			}
		}

		cacheStore.HSet("user:health_factor", userAddress, healthFactor.String())
	}
}
