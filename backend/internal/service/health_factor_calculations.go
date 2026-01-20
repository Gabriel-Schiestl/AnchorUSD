package service

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/domain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

type healthFactorCalculationService struct {
	Store     storage.ICacheStore
	PriceFeed external.IPriceFeedAPI
}

func NewHealthFactorCalculationService(store storage.ICacheStore, priceFeed external.IPriceFeedAPI) *healthFactorCalculationService {
	return &healthFactorCalculationService{
		Store:     store,
		PriceFeed: priceFeed,
	}
}

func (s *healthFactorCalculationService) CalculateMint(ctx context.Context, req model.CalculateMintRequest) (model.HealthFactorProjection, error) {
	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		debtStr = "0"
	}

	collateralUSD := new(big.Int)
	collateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	mintAmount := new(big.Int)
	mintAmount.SetString(req.MintAmount, 10)

	newDebt := new(big.Int).Add(currentDebt, mintAmount)

	healthFactorAfter := domain.CalculateHealthFactorAfterMint(collateralUSD, currentDebt, mintAmount)

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            newDebt.String(),
		NewCollateralValue: collateralUSD.String(),
	}, nil
}

func (s *healthFactorCalculationService) CalculateBurn(ctx context.Context, req model.CalculateBurnRequest) (model.HealthFactorProjection, error) {
	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		debtStr = "0"
	}

	collateralUSD := new(big.Int)
	collateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	burnAmount := new(big.Int)
	burnAmount.SetString(req.BurnAmount, 10)

	newDebt := new(big.Int).Sub(currentDebt, burnAmount)
	if newDebt.Sign() < 0 {
		newDebt = big.NewInt(0)
	}

	healthFactorAfter := domain.CalculateHealthFactorAfterBurn(collateralUSD, currentDebt, burnAmount)

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            newDebt.String(),
		NewCollateralValue: collateralUSD.String(),
	}, nil
}

func (s *healthFactorCalculationService) CalculateDeposit(ctx context.Context, req model.CalculateDepositRequest) (model.HealthFactorProjection, error) {
	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		debtStr = "0"
	}

	currentCollateralUSD := new(big.Int)
	currentCollateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	depositAmount := new(big.Int)
	depositAmount.SetString(req.DepositAmount, 10)

	tokenName := getTokenNameByAddress(req.TokenAddress)
	priceStr, err := getTokenPrice(s.PriceFeed, tokenName)
	if err != nil {
		return model.HealthFactorProjection{}, err
	}

	depositAmountUSD, err := domain.GetTokenAmountInUSD(depositAmount, priceStr)
	if err != nil {
		return model.HealthFactorProjection{}, err
	}

	newCollateralUSD := new(big.Int).Add(currentCollateralUSD, depositAmountUSD)

	healthFactorAfter := domain.CalculateHealthFactorAfterDeposit(currentCollateralUSD, currentDebt, depositAmountUSD)

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            currentDebt.String(),
		NewCollateralValue: newCollateralUSD.String(),
	}, nil
}

func getTokenNameByAddress(tokenAddress string) string {
	for name, address := range constants.CollateralTokens {
		if address == tokenAddress {
			return name
		}
	}
	return ""
}

func getTokenPrice(priceFeed external.IPriceFeedAPI, tokenName string) (string, error) {
	switch tokenName {
	case "ETH":
		return priceFeed.GetEthUsdPrice()
	case "BTC":
		return priceFeed.GetBtcUsdPrice()
	default:
		return "0", nil
	}
}
