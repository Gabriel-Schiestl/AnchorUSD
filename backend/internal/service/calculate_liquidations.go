package service

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

func CalculateLiquidations(priceFeed external.IPriceFeedAPI, cacheStore storage.ICacheStore) error {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting liquidation calculations")

	totalUSDCollateralByUser := make(map[string]*big.Int)

	logger.Debug().Msg("Fetching BTC price")
	btcPrice, err := priceFeed.GetBtcUsdPrice()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get BTC price")
		return err
	}
	logger.Debug().Str("btc_price", btcPrice).Msg("BTC price fetched")

	logger.Debug().Msg("Fetching ETH price")
	ethPrice, err := priceFeed.GetEthUsdPrice()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get ETH price")
		return err
	}
	logger.Debug().Str("eth_price", ethPrice).Msg("ETH price fetched")

	for name, address := range constants.CollateralTokens {
		switch name {
		case "BTC":
			logger.Debug().Str("token", "BTC").Str("address", address).Msg("Processing BTC collateral")
			collateralSupplyByUser, err := cacheStore.HGetAll("collateral:"+address)
			if err != nil {
				logger.Error().Err(err).Str("token", "BTC").Msg("Failed to get BTC collateral supply by user")
				continue
			}

			updateBTCCollateralMetrics(collateralSupplyByUser, btcPrice, totalUSDCollateralByUser)
			logger.Debug().Int("users_count", len(collateralSupplyByUser)).Msg("BTC collateral metrics updated")

		case "ETH":
			logger.Debug().Str("token", "ETH").Str("address", address).Msg("Processing ETH collateral")
			collateralSupplyByUser, err := cacheStore.HGetAll("collateral:" + address)
			if err != nil {
				logger.Error().Err(err).Str("token", "ETH").Msg("Failed to get ETH collateral supply by user")
				continue
			}

			updateETHCollateralMetrics(collateralSupplyByUser, ethPrice, totalUSDCollateralByUser)
			logger.Debug().Int("users_count", len(collateralSupplyByUser)).Msg("ETH collateral metrics updated")
		}
	}

	totalCollateralSupply := new(big.Int)
	logger.Debug().Int("total_users", len(totalUSDCollateralByUser)).Msg("Updating liquidation health factors")
	updateLiquidationHealthFactors(totalCollateralSupply, cacheStore, totalUSDCollateralByUser)

	err = cacheStore.HSet("collateral", "total_supply", totalCollateralSupply.String())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to set total collateral supply")
		return err
	}

	logger.Info().Str("total_collateral_supply", totalCollateralSupply.String()).Msg("Liquidation calculations completed successfully")
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
