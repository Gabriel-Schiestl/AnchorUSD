package service

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/common"
)

func UpdateMetrics(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting metrics update process")

	totalSupply := big.NewInt(0)
	var ethPrice string
	var btcPrice string
	var ethAddr, btcAddr string

	logger.Debug().Msg("Fetching USD prices and addresses for collateral tokens")
	getUsdPriceAndAddress(priceFeed, &ethPrice, &btcPrice, &ethAddr, &btcAddr)
	logger.Debug().Str("eth_price", ethPrice).Str("btc_price", btcPrice).Msg("Prices fetched successfully")

	logger.Info().Msg("Setting coin metrics")
	setCoinMetric(cacheStore, &totalSupply)
	logger.Info().Str("total_supply", totalSupply.String()).Msg("Coin metrics updated")

	logger.Info().Msg("Setting collateral metrics")
	setCollateralMetric(cacheStore, ethPrice, btcPrice, ethAddr, btcAddr)
	logger.Info().Msg("Collateral metrics updated")

	logger.Info().Msg("Metrics update process completed")
}

func getUsdPriceAndAddress(priceFeed external.IPriceFeedAPI, ethPrice *string, btcPrice *string, ethAddr *string, btcAddr *string) {
	logger := utils.GetLogger()
	for name, address := range constants.CollateralTokens {
		switch name {
		case "ETH":
			priceStr, err := priceFeed.GetEthUsdPrice()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to fetch ETH price")
				continue
			}
			*ethPrice = priceStr
			*ethAddr = address
			logger.Debug().Str("price", priceStr).Str("address", address).Msg("ETH price and address loaded")
		case "BTC":
			priceStr, err := priceFeed.GetBtcUsdPrice()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to fetch BTC price")
				continue
			}
			*btcPrice = priceStr
			*btcAddr = address
			logger.Debug().Str("price", priceStr).Str("address", address).Msg("BTC price and address loaded")
		}
	}
}

func setCoinMetric(cacheStore storage.ICacheStore, totalSupply **big.Int) {
	logger := utils.GetLogger()
	logger.Debug().Msg("Fetching total minted by user")

	totalMintedByUser, err := storage.GetCoinStore().GetTotalMintedGroupingByUser(context.Background())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total minted by user")
		return
	}
	logger.Debug().Int("users_count", len(totalMintedByUser)).Msg("Total minted by user fetched")

	totalBurnedByUser, err := storage.GetCoinStore().GetTotalBurnedGroupingByUser(context.Background())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total burned by user")
		return
	}
	logger.Debug().Int("users_count", len(totalBurnedByUser)).Msg("Total burned by user fetched")

	for user, minted := range totalMintedByUser {
		burned, exists := totalBurnedByUser[user]
		if exists {
			minted.Sub(minted, burned)
		}
		(*totalSupply).Add(*totalSupply, minted)
		if minted.Cmp(big.NewInt(0)) < 0 {
			logger.Warn().Str("user", user).Str("calculated_debt", minted.String()).Msg("Negative debt calculated, setting to 0")
			minted = big.NewInt(0)
		}
		cacheStore.HSet("user:debt", user, minted.String())
		logger.Debug().Str("user", user).Str("debt", minted.String()).Msg("User debt updated in cache")
	}

	cacheStore.HSet("coin", "total_supply", (*totalSupply).String())
	logger.Info().Str("total_supply", (*totalSupply).String()).Msg("Total coin supply updated in cache")
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
