package service

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/ethereum/go-ethereum/common"
)

func UpdateMetrics(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	totalSupply := big.NewInt(0)
	var ethPrice string
	var btcPrice string
	var ethAddr, btcAddr string

	getUsdPriceAndAddress(priceFeed, &ethPrice, &btcPrice, &ethAddr, &btcAddr)

	setCoinMetric(cacheStore, &totalSupply)

	setCollateralMetric(cacheStore, ethPrice, btcPrice, ethAddr, btcAddr)
}

func getUsdPriceAndAddress(priceFeed external.IPriceFeedAPI, ethPrice *string, btcPrice *string, ethAddr *string, btcAddr *string) {
	for name, address := range constants.CollateralTokens {
		switch name {
		case "ETH":
			priceStr, err := priceFeed.GetEthUsdPrice()
			if err != nil {
				continue
			}
			*ethPrice = priceStr
			*ethAddr = address
		case "BTC":
			priceStr, err := priceFeed.GetBtcUsdPrice()
			if err != nil {
				continue
			}
			*btcPrice = priceStr
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
		if minted.Cmp(big.NewInt(0)) < 0 {
			minted = big.NewInt(0)
		}
		cacheStore.HSet("user:debt", user, minted.String())
	}

	cacheStore.HSet("coin", "total_supply", (*totalSupply).String())
}

func setCollateralMetric(cacheStore storage.ICacheStore, ethPrice, btcPrice string, ethAddr, btcAddr string) {
	totalCollateral := big.NewInt(0)
	
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

			debt, err := cacheStore.HGet("user:debt", user)
			if err != nil {
				continue
			}

			debtBigInt := big.NewInt(0)
			debtBigInt.SetString(debt, 10)

			healthFactor := domain.CalculateHealthFactor(collateralUsd, debtBigInt)

			cacheStore.HSet("user:health_factor", user, healthFactor.String())

			cacheStore.HAdd("user:collateral_usd", user, collateralUsd)
		}
	}

	cacheStore.HSet("collateral", "total_supply", totalCollateral.String())
}

func getCollateralUSD(collateralType string, deposited *big.Int, ethPrice, btcPrice string, ethAddr, btcAddr string) *big.Int {
	if deposited == nil || deposited.Sign() == 0 {
        return big.NewInt(0)
    }
	fmt.Println("Calculating USD value for collateral type:", collateralType, " with deposited amount:", deposited.String())
    cAddr := common.HexToAddress(collateralType)
    ethA := common.HexToAddress(ethAddr)
    btcA := common.HexToAddress(btcAddr)
	
	res := big.NewInt(0)
	var err error

    if ethA != (common.Address{}) && cAddr == ethA {
		res, err = domain.GetTokenAmountInUSD(deposited, ethPrice)
		if err != nil {
			return big.NewInt(0)
		}
    }

    if btcA != (common.Address{}) && cAddr == btcA {
		res, err = domain.GetTokenAmountInUSD(deposited, btcPrice)
		if err != nil {
			return big.NewInt(0)
		}
    }

    return res
}
