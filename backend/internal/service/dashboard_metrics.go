package service

import (
	"context"
	"math/big"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/metrics"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type dashboardMetricsService struct {
	Store     storage.ICacheStore
	PriceFeed external.IPriceFeedAPI
}

func NewDashboardMetricsService(store storage.ICacheStore, priceFeed external.IPriceFeedAPI) *dashboardMetricsService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing dashboard metrics service")
	return &dashboardMetricsService{
		Store:     store,
		PriceFeed: priceFeed,
	}
}

func (s *dashboardMetricsService) GetDashboardMetrics(ctx context.Context) (model.DashboardMetrics, error) {
	start := time.Now()
	defer func() {
		metrics.RecordOperation("get_dashboard_metrics", time.Since(start).Seconds())
	}()

	logger := utils.GetLogger()
	logger.Info().Msg("Starting dashboard metrics aggregation")

	liquidatableUsers, err := s.getLiquidatableUsers()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get liquidatable users, using empty list")
		liquidatableUsers = []model.LiquidatableUser{}
	}
	logger.Debug().Int("count", len(liquidatableUsers)).Msg("Liquidatable users retrieved")
	
	metrics.LiquidatableUsers.Set(float64(len(liquidatableUsers)))

	totalCollateral, err := s.getTotalCollateral()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get total collateral, using default")
		totalCollateral = model.TotalCollateral{Value: "0", Breakdown: []model.CollateralBreakdown{}}
	}
	logger.Debug().Str("value", totalCollateral.Value).Msg("Total collateral retrieved")

	stableSupply, err := s.getStableSupply()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get stable supply, using default")
		stableSupply = model.StableSupply{Total: "0", Circulating: "0", Backing: 0}
	}
	logger.Debug().Str("total", stableSupply.Total).Str("circulating", stableSupply.Circulating).Float64("backing", stableSupply.Backing).Msg("Stable supply retrieved")
	
	metrics.BackingPercentage.Set(stableSupply.Backing)

	protocolHealth, err := s.getProtocolHealth()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get protocol health, using default")
		protocolHealth = model.ProtocolHealth{
			AverageHealthFactor:    0,
			UsersAtRisk:            0,
			TotalUsers:             0,
			CollateralizationRatio: 0,
		}
	}
	logger.Debug().Float64("avg_health_factor", protocolHealth.AverageHealthFactor).Int("users_at_risk", protocolHealth.UsersAtRisk).Int("total_users", protocolHealth.TotalUsers).Msg("Protocol health retrieved")
	
	metrics.AverageHealthFactor.Set(protocolHealth.AverageHealthFactor)
	metrics.UsersAtRisk.Set(float64(protocolHealth.UsersAtRisk))
	metrics.ActiveUsersTotal.Set(float64(protocolHealth.TotalUsers))

	logger.Info().Msg("Dashboard metrics aggregation completed successfully")

	return model.DashboardMetrics{
		LiquidatableUsers: liquidatableUsers,
		TotalCollateral:   totalCollateral,
		StableSupply:      stableSupply,
		ProtocolHealth:    protocolHealth,
	}, nil
}

func (s *dashboardMetricsService) getLiquidatableUsers() ([]model.LiquidatableUser, error) {
	liquidatableMap, err := s.Store.HGetAll("liquidatable")
	if err != nil {
		return []model.LiquidatableUser{}, nil
	}

	users := make([]model.LiquidatableUser, 0, len(liquidatableMap))

	for address, healthFactorStr := range liquidatableMap {
		collateralUsd, err := s.Store.HGet("user:collateral_usd", address)
		if err != nil {
			collateralUsd = "0"
		}

		debtUsd, err := s.Store.HGet("user:debt", address)
		if err != nil {
			debtUsd = "0"
		}

		debtBigInt := new(big.Int)
		debtBigInt.SetString(debtUsd, 10)
		liquidationAmount := domain.CalculateLiquidationAmount(debtBigInt)

		users = append(users, model.LiquidatableUser{
			Address:           address,
			HealthFactor:      healthFactorStr,
			CollateralUsd:     collateralUsd,
			DebtUsd:           debtUsd,
			LiquidationAmount: liquidationAmount.String(),
		})
	}

	return users, nil
}

func (s *dashboardMetricsService) getTotalCollateral() (model.TotalCollateral, error) {
	totalCollateralStr, err := s.Store.HGet("collateral", "total_supply")
	if err != nil {
		totalCollateralStr = "0"
	}

	totalCollateralBigInt := new(big.Int)
	totalCollateralBigInt.SetString(totalCollateralStr, 10)

	breakdown := []model.CollateralBreakdown{}

	ethPrice, _ := s.PriceFeed.GetEthUsdPrice()
	btcPrice, _ := s.PriceFeed.GetBtcUsdPrice()

	for name, tokenAddress := range constants.CollateralTokens {
		collateralKey := "collateral:" + tokenAddress

		usersCollateral, err := s.Store.HGetAll(collateralKey)
		if err != nil {
			continue
		}

		totalAmount := new(big.Int)
		for _, amountStr := range usersCollateral {
			amount := new(big.Int)
			amount.SetString(amountStr, 10)
			totalAmount.Add(totalAmount, amount)
		}

		var priceStr string
		switch name {
		case "ETH":
			priceStr = ethPrice
		case "BTC":
			priceStr = btcPrice
		default:
			continue
		}

		valueUsd, err := domain.GetTokenAmountInUSD(totalAmount, priceStr)
		if err != nil {
			valueUsd = big.NewInt(0)
		}

		percentage := domain.PercentageOf(valueUsd, totalCollateralBigInt)

		breakdown = append(breakdown, model.CollateralBreakdown{
			Asset:      name,
			Amount:     totalAmount.String(),
			ValueUsd:   valueUsd.String(),
			Percentage: percentage,
		})
	}

	return model.TotalCollateral{
		Value:     totalCollateralStr,
		Breakdown: breakdown,
	}, nil
}

func (s *dashboardMetricsService) getStableSupply() (model.StableSupply, error) {
	totalSupplyStr, err := s.Store.HGet("coin", "total_supply")
	if err != nil {
		totalSupplyStr = "0"
	}

	usersDebt, err := s.Store.HGetAll("user:debt")
	if err != nil {
		return model.StableSupply{
			Total:       totalSupplyStr,
			Circulating: totalSupplyStr,
			Backing:     0,
		}, nil
	}

	circulatingSupply := new(big.Int)
	for _, debtStr := range usersDebt {
		debt := new(big.Int)
		debt.SetString(debtStr, 10)
		circulatingSupply.Add(circulatingSupply, debt)
	}

	totalCollateralStr, _ := s.Store.HGet("collateral", "total_supply")
	totalCollateral := new(big.Int)
	totalCollateral.SetString(totalCollateralStr, 10)

	backing := domain.CalculateBackingPercentage(totalCollateral, circulatingSupply)

	return model.StableSupply{
		Total:       totalSupplyStr,
		Circulating: circulatingSupply.String(),
		Backing:     backing,
	}, nil
}

func (s *dashboardMetricsService) getProtocolHealth() (model.ProtocolHealth, error) {
	usersHealthFactors, err := s.Store.HGetAll("user:health_factor")
	if err != nil {
		return model.ProtocolHealth{}, err
	}

	totalUsers := len(usersHealthFactors)
	if totalUsers == 0 {
		return model.ProtocolHealth{
			AverageHealthFactor:    0,
			UsersAtRisk:            0,
			TotalUsers:             0,
			CollateralizationRatio: 0,
		}, nil
	}

	sumHealthFactor := new(big.Int)
	usersAtRisk := 0

	for _, hfStr := range usersHealthFactors {
		hf := new(big.Int)
		hf.SetString(hfStr, 10)
		sumHealthFactor.Add(sumHealthFactor, hf)

		if domain.IsAtRisk(hf) {
			usersAtRisk++
		}
	}

	avgHealthFactorFloat := domain.AverageHealthFactor(sumHealthFactor, totalUsers)

	totalCollateralStr, _ := s.Store.HGet("collateral", "total_supply")
	totalCollateral := new(big.Int)
	totalCollateral.SetString(totalCollateralStr, 10)

	totalDebtStr, _ := s.Store.HGet("coin", "total_supply")
	totalDebt := new(big.Int)
	totalDebt.SetString(totalDebtStr, 10)

	collateralizationRatio := domain.CollateralizationRatio(totalCollateral, totalDebt)

	return model.ProtocolHealth{
		AverageHealthFactor:    avgHealthFactorFloat,
		UsersAtRisk:            usersAtRisk,
		TotalUsers:             totalUsers,
		CollateralizationRatio: collateralizationRatio,
	}, nil
}
