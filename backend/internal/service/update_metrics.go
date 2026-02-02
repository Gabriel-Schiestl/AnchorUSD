package service

import (
	"context"
	"math/big"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/metrics"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/common"
)

var totalSupply = big.NewInt(0)
var cacheStore storage.ICacheStore
var logger = utils.GetLogger()
var ethPrice string
var btcPrice string
var ethAddr string
var btcAddr string

func UpdateMetrics(IcacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	start := time.Now()
	defer func() {
		metrics.RecordOperation("update_metrics", time.Since(start).Seconds())
	}()

	cacheStore = IcacheStore
	logger.Info().Msg("Starting metrics update process")

	logger.Debug().Msg("Fetching USD prices and addresses for collateral tokens")
	getUsdPriceAndAddress(priceFeed)
	logger.Debug().Str("eth_price", ethPrice).Str("btc_price", btcPrice).Msg("Prices fetched successfully")

	logger.Info().Msg("Setting coin metrics")
	setCoinMetric()
	logger.Info().Str("total_supply", totalSupply.String()).Msg("Coin metrics updated")

	logger.Info().Msg("Setting collateral metrics")
	setCollateralMetric()
	logger.Info().Msg("Collateral metrics updated")

	logger.Info().Msg("Metrics update process completed")
}

func getUsdPriceAndAddress(priceFeed external.IPriceFeedAPI) {
	logger := utils.GetLogger()
	for name, address := range constants.CollateralTokens {
		switch name {
		case "ETH":
			priceStr, err := priceFeed.GetEthUsdPrice()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to fetch ETH price")
				metrics.PriceFeedErrors.WithLabelValues("ETH").Inc()
				continue
			}
			ethPrice = priceStr
			ethAddr = address
			metrics.TokenPriceUSD.WithLabelValues("ETH").Set(parseFloat64(priceStr))
			logger.Debug().Str("price", priceStr).Str("address", address).Msg("ETH price and address loaded")
		case "BTC":
			priceStr, err := priceFeed.GetBtcUsdPrice()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to fetch BTC price")
				metrics.PriceFeedErrors.WithLabelValues("BTC").Inc()
				continue
			}
			btcPrice = priceStr
			btcAddr = address
			metrics.TokenPriceUSD.WithLabelValues("BTC").Set(parseFloat64(priceStr))
			logger.Debug().Str("price", priceStr).Str("address", address).Msg("BTC price and address loaded")
		}
	}
}

func setCoinMetric() {
	logger.Debug().Msg("Fetching total minted by user")

	err := storage.GetCoinStore().IterateTotalMintedGroupingByUser(context.Background(), 500, iteratorTotalMintedByUserCallback)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total minted by user")
		return
	}
}

func iteratorTotalMintedByUserCallback(totalMintedByUser map[string]*big.Int) error {
	logger.Debug().Int("users_count", len(totalMintedByUser)).Msg("Total minted by user fetched")

	users := make([]string, 0, len(totalMintedByUser))
	for user := range totalMintedByUser {
		users = append(users, user)
	}

	logger.Debug().Msg("Fetching total burned by user")

	totalBurnedByUser, err := storage.GetCoinStore().GetTotalBurnedGroupingByUser(context.Background(), users)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total burned by user")
		return err
	}
	logger.Debug().Int("users_count", len(totalBurnedByUser)).Msg("Total burned by user fetched")

	for user, minted := range totalMintedByUser {
		burned, exists := totalBurnedByUser[user]
		if exists {
			minted.Sub(minted, burned)
		}
		totalSupply.Add(totalSupply, minted)
		if minted.Cmp(big.NewInt(0)) < 0 {
			logger.Warn().Str("user", user).Str("calculated_debt", minted.String()).Msg("Negative debt calculated, setting to 0")
			minted = big.NewInt(0)
		}
		cacheStore.HSet("user:debt", user, minted.String())
		logger.Debug().Str("user", user).Str("debt", minted.String()).Msg("User debt updated in cache")
	}

	cacheStore.HSet("coin", "total_supply", totalSupply.String())
	logger.Info().Str("total_supply", totalSupply.String()).Msg("Total coin supply updated in cache")
	
	metrics.AUSDTotalSupply.Set(parseFloat64(totalSupply.String()))
	
	return nil
}

func setCollateralMetric() {
	logger.Debug().Msg("Fetching total deposited by user")

	err := storage.GetCollateralStore().IterateTotalDepositedGroupingByUser(context.Background(), 500, iteratorTotalDepositedByUserCallback)
	if err != nil {
		return
	}

	logger.Debug().Msg("Collateral metrics updated successfully")
}

func iteratorTotalDepositedByUserCallback(totalDepositedByUser map[string]map[string]*big.Int) error {
	logger.Debug().Int("users_count", len(totalDepositedByUser)).Msg("Total deposited by user fetched")
	totalCollateral := big.NewInt(0)

	users := make([]string, 0, len(totalDepositedByUser))
	for user := range totalDepositedByUser {
		users = append(users, user)
	}

	logger.Debug().Msg("Fetching total redeemed by user")

	totalRedeemedByUser, err := storage.GetCollateralStore().GetTotalCollateralRedeemedGroupingByUser(context.Background(), users)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total redeemed by user")
		return err
	}

	logger.Debug().Int("users_count", len(totalRedeemedByUser)).Msg("Total redeemed by user fetched")

	for user, collateralMap := range totalDepositedByUser {
		for collateralType, deposited := range collateralMap {
			redeemedMap, exists := totalRedeemedByUser[user]
			if exists {
				redeemed, redeemedExists := redeemedMap[collateralType]
				if redeemedExists {
					logger.Debug().Str("user", user).Str("collateral_type", collateralType).Str("deposited", deposited.String()).Str("redeemed", redeemed.String()).Msg("Calculating net deposited collateral")
					deposited.Sub(deposited, redeemed)
				}
			}
			logger.Debug().Str("user", user).Str("collateral_type", collateralType).Str("net_deposited", deposited.String()).Msg("Net deposited collateral calculated")
			
			cacheStore.HAdd("collateral:"+collateralType, user, deposited)

			logger.Debug().Str("user", user).Str("collateral_type", collateralType).Str("deposited", deposited.String()).Msg("User collateral updated in cache")

			var collateralUsd *big.Int
			collateralUsd = getCollateralUSD(collateralType, deposited, ethPrice, btcPrice, ethAddr, btcAddr)

			logger.Debug().Str("user", user).Str("collateral_type", collateralType).Str("collateral_usd", collateralUsd.String()).Msg("Collateral USD value calculated")

			totalCollateral.Add(totalCollateral, collateralUsd)

			debt, err := cacheStore.HGet("user:debt", user)
			if err != nil {
				logger.Error().Err(err).Str("user", user).Msg("Failed to get user debt from cache")
				continue
			}

			logger.Debug().Str("user", user).Str("debt", debt).Msg("User debt fetched from cache")

			debtBigInt := big.NewInt(0)
			debtBigInt.SetString(debt, 10)

			healthFactor := domain.CalculateHealthFactor(collateralUsd, debtBigInt)

			logger.Debug().Str("user", user).Str("health_factor", healthFactor.String()).Msg("Health factor calculated")

			cacheStore.HSet("user:health_factor", user, healthFactor.String())

			logger.Debug().Str("user", user).Str("health_factor", healthFactor.String()).Msg("User health factor updated in cache")

			cacheStore.HAdd("user:collateral_usd", user, collateralUsd)

			logger.Debug().Str("user", user).Str("collateral_usd", collateralUsd.String()).Msg("User collateral USD value updated in cache")
			
			tokenName := getTokenNameByAddress(collateralType)
			if tokenName != "" {
				metrics.TotalCollateralUSD.WithLabelValues(tokenName).Set(parseFloat64(collateralUsd.String()))
			}
		}
	}

	cacheStore.HSet("collateral", "total_supply", totalCollateral.String())
	
	metrics.CollateralizationRatio.Set(calculateCollateralizationRatio(totalCollateral, totalSupply))
	
	return nil
}

func parseFloat64(value string) float64 {
	bigIntValue := new(big.Int)
	bigIntValue.SetString(value, 10)
	
	bigFloatValue := new(big.Float).SetInt(bigIntValue)
	result, _ := bigFloatValue.Float64()
	
	return result / 1e18
}

func calculateCollateralizationRatio(totalCollateral, totalSupply *big.Int) float64 {
	if totalSupply.Sign() == 0 {
		return 0
	}
	
	collateralFloat := parseFloat64(totalCollateral.String())
	supplyFloat := parseFloat64(totalSupply.String())
	
	if supplyFloat == 0 {
		return 0
	}
	
	return (collateralFloat / supplyFloat) * 100
}

func getCollateralUSD(collateralType string, deposited *big.Int, ethPrice, btcPrice string, ethAddr, btcAddr string) *big.Int {
	if deposited == nil || deposited.Sign() == 0 {
        return big.NewInt(0)
    }
	logger.Debug().Str("collateral_type", collateralType).Str("deposited", deposited.String()).Msg("Calculating collateral USD value")
	
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
