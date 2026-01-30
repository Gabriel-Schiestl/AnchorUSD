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

type healthFactorCalculationService struct {
	Store     storage.ICacheStore
	PriceFeed external.IPriceFeedAPI
}

func NewHealthFactorCalculationService(store storage.ICacheStore, priceFeed external.IPriceFeedAPI) *healthFactorCalculationService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing health factor calculation service")
	return &healthFactorCalculationService{
		Store:     store,
		PriceFeed: priceFeed,
	}
}

func (s *healthFactorCalculationService) CalculateMint(ctx context.Context, req model.CalculateMintRequest) (model.HealthFactorProjection, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", req.Address).Str("amount", req.MintAmount).Msg("Calculating health factor projection for mint operation")

	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("Collateral USD not found, defaulting to 0")
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("User debt not found, defaulting to 0")
		debtStr = "0"
	}

	collateralUSD := new(big.Int)
	collateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	mintAmount := new(big.Int)
	mintAmount.SetString(req.MintAmount, 10)

	newDebt := new(big.Int).Add(currentDebt, mintAmount)

	logger.Debug().Str("user", req.Address).Str("collateral_usd", collateralUSDStr).Str("current_debt", debtStr).Str("mint_amount", req.MintAmount).Msg("Calculating health factor after mint")

	healthFactorAfter := domain.CalculateHealthFactorAfterMint(collateralUSD, currentDebt, mintAmount)

	logger.Info().Str("user", req.Address).Str("health_factor_after", healthFactorAfter.String()).Str("new_debt", newDebt.String()).Msg("Mint health factor projection calculated successfully")

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            newDebt.String(),
		NewCollateralValue: collateralUSD.String(),
	}, nil
}

func (s *healthFactorCalculationService) CalculateBurn(ctx context.Context, req model.CalculateBurnRequest) (model.HealthFactorProjection, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", req.Address).Str("amount", req.BurnAmount).Msg("Calculating health factor projection for burn operation")

	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("Collateral USD not found, defaulting to 0")
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("User debt not found, defaulting to 0")
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
		logger.Warn().Str("user", req.Address).Str("current_debt", debtStr).Str("burn_amount", req.BurnAmount).Msg("Burn amount exceeds debt, setting new debt to 0")
		newDebt = big.NewInt(0)
	}

	logger.Debug().Str("user", req.Address).Str("collateral_usd", collateralUSDStr).Str("current_debt", debtStr).Str("burn_amount", req.BurnAmount).Msg("Calculating health factor after burn")

	healthFactorAfter := domain.CalculateHealthFactorAfterBurn(collateralUSD, currentDebt, burnAmount)

	logger.Info().Str("user", req.Address).Str("health_factor_after", healthFactorAfter.String()).Str("new_debt", newDebt.String()).Msg("Burn health factor projection calculated successfully")

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            newDebt.String(),
		NewCollateralValue: collateralUSD.String(),
	}, nil
}

func (s *healthFactorCalculationService) CalculateDeposit(ctx context.Context, req model.CalculateDepositRequest) (model.HealthFactorProjection, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", req.Address).Str("token", req.TokenAddress).Str("amount", req.DepositAmount).Msg("Calculating health factor projection for deposit operation")

	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("Collateral USD not found, defaulting to 0")
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("User debt not found, defaulting to 0")
		debtStr = "0"
	}

	currentCollateralUSD := new(big.Int)
	currentCollateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	depositAmount := new(big.Int)
	depositAmount.SetString(req.DepositAmount, 10)

	tokenName := getTokenNameByAddress(req.TokenAddress)
	logger.Debug().Str("token_address", req.TokenAddress).Str("token_name", tokenName).Msg("Token identified")

	priceStr, err := getTokenPrice(s.PriceFeed, tokenName)
	if err != nil {
		logger.Error().Err(err).Str("token_name", tokenName).Msg("Failed to get token price")
		return model.HealthFactorProjection{}, err
	}
	logger.Debug().Str("token_name", tokenName).Str("price_usd", priceStr).Msg("Token price fetched")

	depositAmountUSD, err := domain.GetTokenAmountInUSD(depositAmount, priceStr)
	if err != nil {
		logger.Error().Err(err).Str("deposit_amount", req.DepositAmount).Str("price", priceStr).Msg("Failed to convert deposit amount to USD")
		return model.HealthFactorProjection{}, err
	}

	newCollateralUSD := new(big.Int).Add(currentCollateralUSD, depositAmountUSD)

	logger.Debug().Str("user", req.Address).Str("current_collateral_usd", collateralUSDStr).Str("deposit_usd", depositAmountUSD.String()).Str("new_collateral_usd", newCollateralUSD.String()).Msg("Calculating health factor after deposit")

	healthFactorAfter := domain.CalculateHealthFactorAfterDeposit(currentCollateralUSD, currentDebt, depositAmountUSD)

	logger.Info().Str("user", req.Address).Str("token", tokenName).Str("health_factor_after", healthFactorAfter.String()).Str("new_collateral_value", newCollateralUSD.String()).Msg("Deposit health factor projection calculated successfully")

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            currentDebt.String(),
		NewCollateralValue: newCollateralUSD.String(),
	}, nil
}

func (s *healthFactorCalculationService) CalculateRedeem(ctx context.Context, req model.CalculateRedeemRequest) (model.HealthFactorProjection, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", req.Address).Str("token", req.TokenAddress).Str("amount", req.RedeemAmount).Msg("Calculating health factor projection for redeem operation")

	collateralUSDStr, err := s.Store.HGet("user:collateral_usd", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("Collateral USD not found, defaulting to 0")
		collateralUSDStr = "0"
	}

	debtStr, err := s.Store.HGet("user:debt", req.Address)
	if err != nil {
		logger.Debug().Err(err).Str("user", req.Address).Msg("User debt not found, defaulting to 0")
		debtStr = "0"
	}

	currentCollateralUSD := new(big.Int)
	currentCollateralUSD.SetString(collateralUSDStr, 10)

	currentDebt := new(big.Int)
	currentDebt.SetString(debtStr, 10)

	redeemAmount := new(big.Int)
	redeemAmount.SetString(req.RedeemAmount, 10)

	tokenName := getTokenNameByAddress(req.TokenAddress)
	logger.Debug().Str("token_address", req.TokenAddress).Str("token_name", tokenName).Msg("Token identified")

	priceStr, err := getTokenPrice(s.PriceFeed, tokenName)
	if err != nil {
		logger.Error().Err(err).Str("token_name", tokenName).Msg("Failed to get token price")
		return model.HealthFactorProjection{}, err
	}
	logger.Debug().Str("token_name", tokenName).Str("price_usd", priceStr).Msg("Token price fetched")

	redeemAmountUSD, err := domain.GetTokenAmountInUSD(redeemAmount, priceStr)
	if err != nil {
		logger.Error().Err(err).Str("redeem_amount", req.RedeemAmount).Str("price", priceStr).Msg("Failed to convert redeem amount to USD")
		return model.HealthFactorProjection{}, err
	}

	newCollateralUSD := new(big.Int).Sub(currentCollateralUSD, redeemAmountUSD)

	logger.Debug().Str("user", req.Address).Str("current_collateral_usd", collateralUSDStr).Str("redeem_usd", redeemAmountUSD.String()).Str("new_collateral_usd", newCollateralUSD.String()).Msg("Calculating health factor after redeem")

	healthFactorAfter := domain.CalculateHealthFactorAfterDeposit(currentCollateralUSD, currentDebt, redeemAmountUSD)

	logger.Info().Str("user", req.Address).Str("token", tokenName).Str("health_factor_after", healthFactorAfter.String()).Str("new_collateral_value", newCollateralUSD.String()).Msg("Deposit health factor projection calculated successfully")

	return model.HealthFactorProjection{
		HealthFactorAfter:  healthFactorAfter.String(),
		NewDebt:            currentDebt.String(),
		NewCollateralValue: newCollateralUSD.String(),
	}, nil
}

func getTokenNameByAddress(tokenAddress string) string {
	logger := utils.GetLogger()
	for name, address := range constants.CollateralTokens {
		if address == tokenAddress {
			logger.Debug().Str("address", tokenAddress).Str("name", name).Msg("Token name found")
			return name
		}
	}
	logger.Warn().Str("address", tokenAddress).Msg("Token name not found for address")
	return ""
}

func getTokenPrice(priceFeed external.IPriceFeedAPI, tokenName string) (string, error) {
	logger := utils.GetLogger()
	logger.Debug().Str("token", tokenName).Msg("Fetching token price")

	switch tokenName {
	case "ETH":
		price, err := priceFeed.GetEthUsdPrice()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to fetch ETH price")
		}
		return price, err
	case "BTC":
		price, err := priceFeed.GetBtcUsdPrice()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to fetch BTC price")
		}
		return price, err
	default:
		logger.Warn().Str("token", tokenName).Msg("Unknown token, returning 0 price")
		return "0", nil
	}
}
