package service

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

type userDataService struct {
	Store storage.ICacheStore
}

func NewUserDataService(store storage.ICacheStore) *userDataService {
	return &userDataService{
		Store: store,
	}
}

func (s *userDataService) GetUserData(ctx context.Context, user string) (model.UserData, error) {
	totalDebt, err := s.Store.HGet("user:debt", user)
	if err != nil {
		totalDebt = "0"
	}

	collateralValueUSD, err := s.Store.HGet("user:collateral_usd", user)
	if err != nil {
		collateralValueUSD = "0"
	}

	healthFactor, err := s.Store.HGet("user:health_factor", user)
	if err != nil {
		healthFactor = "0"
	}

	maxMintable := domain.CalculateMaxMintable(collateralValueUSD, totalDebt)

	return model.UserData{
		TotalDebt:           totalDebt,
		CollateralValueUSD:  collateralValueUSD,
		MaxMintable:         maxMintable,
		CurrentHealthFactor: healthFactor,
	}, nil
}

