package processors

import (
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

func ProcessCoin(metric model.Metrics, cacheStore storage.ICacheStore) {
	logger := utils.GetLogger()
	logger.Info().Str("user", metric.UserAddress.Hex()).Str("amount", metric.Amount.String()).Str("operation", string(metric.Operation)).Msg("Processing coin metric")

	amountToChange := getAmountChange(metric)
	logger.Debug().Str("amount_to_change", amountToChange.String()).Msg("Calculated amount to change")

	cacheStore.HAdd("coin", "total_supply", amountToChange)
	logger.Debug().Str("total_supply_change", amountToChange.String()).Msg("Updated total coin supply")

	cacheStore.HAdd("user:debt", metric.UserAddress.Hex(), amountToChange)
	logger.Debug().Str("user", metric.UserAddress.Hex()).Str("debt_change", amountToChange.String()).Msg("Updated user debt")

	collateralUSDValue, err := cacheStore.HGet("user:collateral_usd", metric.UserAddress.Hex())
	if err != nil {
		logger.Error().Err(err).Str("user", metric.UserAddress.Hex()).Msg("Failed to get user collateral USD value")
		return
	}

	debt, err := cacheStore.HGet("user:debt", metric.UserAddress.Hex())
	if err != nil {
		logger.Error().Err(err).Str("user", metric.UserAddress.Hex()).Msg("Failed to get user debt")
		return
	}

	debtBigInt := big.NewInt(0)
	debtBigInt.SetString(debt, 10)

	collateralUSDBigInt := big.NewInt(0)
	collateralUSDBigInt.SetString(collateralUSDValue, 10)

	logger.Debug().Str("user", metric.UserAddress.Hex()).Str("collateral_usd", collateralUSDValue).Str("debt", debt).Msg("Calculating health factor")

	healthFactor := domain.CalculateHealthFactor(collateralUSDBigInt, debtBigInt)

	cacheStore.HSet("user:health_factor", metric.UserAddress.Hex(), healthFactor.String())
	logger.Info().Str("user", metric.UserAddress.Hex()).Str("health_factor", healthFactor.String()).Msg("Coin metric processed and health factor updated")
}
