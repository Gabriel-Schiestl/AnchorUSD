package service

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type userDataService struct {
	Store     storage.ICacheStore
	PriceFeed external.IPriceFeedAPI
}

func NewUserDataService(store storage.ICacheStore, priceFeed external.IPriceFeedAPI) *userDataService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing user data service")
	return &userDataService{
		Store:     store,
		PriceFeed: priceFeed,
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

	collateralAssets := s.fetchCollateralAssets(user)

	collateralDeposited := domain.CalculateCollateralDeposited(collateralAssets)

	userData := model.UserData{
		TotalDebt:           totalDebt,
		CollateralValueUSD:  collateralValueUSD,
		MaxMintable:         maxMintable,
		CurrentHealthFactor: healthFactor,
		CollateralDeposited: collateralDeposited,
	}

	logger.Info().Str("user", user).Str("max_mintable", maxMintable).Int("collateral_assets", len(collateralDeposited)).Msg("User data retrieved successfully")

	return userData, nil
}

func (s *userDataService) fetchCollateralAssets(user string) []domain.CollateralAssetData {
	logger := utils.GetLogger()
	assets := []domain.CollateralAssetData{}

	ethPrice, _ := s.PriceFeed.GetEthUsdPrice()
	btcPrice, _ := s.PriceFeed.GetBtcUsdPrice()

	for name, tokenAddress := range constants.CollateralTokens {
		collateralKey := "collateral:" + tokenAddress

		amountStr, err := s.Store.HGet(collateralKey, user)
		if err != nil {
			logger.Debug().Err(err).Str("user", user).Str("asset", name).Msg("No collateral found for this asset")
			continue
		}

		amount := new(big.Int)
		if _, ok := amount.SetString(amountStr, 10); !ok {
			logger.Warn().Str("user", user).Str("asset", name).Str("amount", amountStr).Msg("Failed to parse collateral amount")
			continue
		}

		var priceStr string
		switch name {
		case "ETH":
			priceStr = ethPrice
		case "BTC":
			priceStr = btcPrice
		default:
			logger.Debug().Str("asset", name).Msg("Unknown asset type, skipping")
			continue
		}

		assets = append(assets, domain.CollateralAssetData{
			Name:     name,
			Amount:   amount,
			PriceUSD: priceStr,
		})
	}

	return assets
}
