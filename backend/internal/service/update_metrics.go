package service

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func UpdateMetrics(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	var totalCollateral *big.Int
	var totalSupply *big.Int
	var ethPrice *big.Int
	var btcPrice *big.Int
	var ethAddr, btcAddr string

	getUsdPriceAndAddress(priceFeed, &ethPrice, &btcPrice, &ethAddr, &btcAddr)

	setCoinMetric(cacheStore, &totalSupply)

	setCollateralMetric(cacheStore, ethPrice, btcPrice, ethAddr, btcAddr, totalCollateral)
}

func getUsdPriceAndAddress(priceFeed external.IPriceFeedAPI, ethPrice **big.Int, btcPrice **big.Int, ethAddr *string, btcAddr *string) {
	for name, address := range constants.CollateralTokens {
		switch name {
		case "ETH":
			priceStr, err := priceFeed.GetEthUsdPrice()
			if err != nil {
				continue
			}
			priceBigInt := big.NewInt(0)
			priceBigInt.SetString(priceStr, 10)

			*ethPrice = new(big.Int).Mul(priceBigInt, constants.PRICE_PRECISION)
			*ethAddr = address
		case "BTC":
			priceStr, err := priceFeed.GetBtcUsdPrice()
			if err != nil {
				continue
			}
			priceBigInt := big.NewInt(0)
			priceBigInt.SetString(priceStr, 10)

			*btcPrice = new(big.Int).Mul(priceBigInt, constants.PRICE_PRECISION)
			*btcAddr = address
		}
	}
}

func setCoinMetric(cacheStore storage.ICacheStore, totalSupply **big.Int) {
	totalMintedByUser, err := storage.GetCoinStore().GetTotalMintedGroupingByUser(context.Background())
	if err != nil {
		return
	}

	totalBurnedByUser, err := storage.GetCoinStore().GetTotalBurnedGroupingByUser(context.Background())
	if err != nil {
		return
	}

	for user, minted := range totalMintedByUser {
		burned, exists := totalBurnedByUser[user]
		if exists {
			minted.Sub(minted, burned)
		}
		(*totalSupply).Add(*totalSupply, minted)
		cacheStore.Set("user:debt:"+user, minted.String(), 0)
	}

	cacheStore.Set("coin:total_supply", (*totalSupply).String(), 0)
}

func setCollateralMetric(cacheStore storage.ICacheStore, ethPrice, btcPrice *big.Int, ethAddr, btcAddr string, totalCollateral *big.Int) {
	totalDepositedByUser, err := storage.GetCollateralStore().GetTotalCollateralDepositedGroupingByUser(context.Background())
	if err != nil {
		return
	}

	totalRedeemedByUser, err := storage.GetCollateralStore().GetTotalCollateralRedeemedGroupingByUser(context.Background())
	if err != nil {
		return
	}

	for user, collateralMap := range totalDepositedByUser {
		for collateralType, deposited := range collateralMap {
			redeemedMap, exists := totalRedeemedByUser[user]
			if exists {
				redeemed, redeemedExists := redeemedMap[collateralType]
				if redeemedExists {
					deposited.Sub(deposited, redeemed)
				}
			}
			cacheStore.HAdd("collateral:"+collateralType, user, deposited)

			var collateralUsd *big.Int
			collateralUsd = getCollateralUSD(collateralType, deposited, ethPrice, btcPrice, ethAddr, btcAddr)

			totalCollateral.Add(totalCollateral, collateralUsd)

			debt, err := cacheStore.Get("user:debt:" + user)
			if err != nil {
				continue
			}

			debtBigInt := big.NewInt(0)
			debtBigInt.SetString(debt, 10)

			healthFactor := domain.CalculateHealthFactor(collateralUsd, debtBigInt)

			cacheStore.Set("user:health_factor:"+user, healthFactor.String(), 0)

			cacheStore.Add("user:collateral_usd:"+user, collateralUsd)
		}
	}

	cacheStore.Set("collateral:total_supply", totalCollateral.String(), 0)
}

func getCollateralUSD(collateralType string, deposited *big.Int, ethPrice, btcPrice *big.Int, ethAddr, btcAddr string) *big.Int {
	var collateralUsd *big.Int
	switch collateralType {
	case ethAddr:
		collateralUsd = big.NewInt(0).Mul(deposited, ethPrice)
		collateralUsd = collateralUsd.Div(collateralUsd, big.NewInt(1e18))
	case btcAddr:
		collateralUsd = big.NewInt(0).Mul(deposited, btcPrice)
		collateralUsd = collateralUsd.Div(collateralUsd, big.NewInt(1e8))
	}
	return collateralUsd
}
