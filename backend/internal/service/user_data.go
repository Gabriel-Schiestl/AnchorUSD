package service

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type userDataService struct {
	Store storage.ICacheStore
}

func NewUserDataService(store storage.ICacheStore) *userDataService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing user data service")
	return &userDataService{
		Store: store,
	}
}

func (s *userDataService) GetUserData(ctx context.Context, user string) (model.UserData, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", user).Msg("Getting user data from cache")

	totalDebt, err := s.Store.HGet("user:debt", user)
	if err != nil {
		logger.Debug().Err(err).Str("user", user).Msg("User debt not found in cache, defaulting to 0")
		totalDebt = "0"
	}

	collateralValueUSD, err := s.Store.HGet("user:collateral_usd", user)
	if err != nil {
		logger.Debug().Err(err).Str("user", user).Msg("User collateral USD not found in cache, defaulting to 0")
		collateralValueUSD = "0"
	}

	healthFactor, err := s.Store.HGet("user:health_factor", user)
	if err != nil {
		logger.Debug().Err(err).Str("user", user).Msg("User health factor not found in cache, defaulting to 0")
		healthFactor = "0"
	}

	logger.Debug().Str("user", user).Str("total_debt", totalDebt).Str("collateral_usd", collateralValueUSD).Str("health_factor", healthFactor).Msg("Calculating max mintable")

	maxMintable := domain.CalculateMaxMintable(collateralValueUSD, totalDebt)

	userData := model.UserData{
		TotalDebt:           totalDebt,
		CollateralValueUSD:  collateralValueUSD,
		MaxMintable:         maxMintable,
		CurrentHealthFactor: healthFactor,
	}

	logger.Info().Str("user", user).Str("max_mintable", maxMintable).Msg("User data retrieved successfully")

	return userData, nil
}

