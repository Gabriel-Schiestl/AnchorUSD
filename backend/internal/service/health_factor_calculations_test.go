package service

import (
	"context"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthFactorCalculationService(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)

	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	assert.NotNil(t, service)
	assert.Equal(t, mockCache, service.Store)
	assert.Equal(t, mockPriceFeed, service.PriceFeed)
}

func TestCalculateMint_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("50000000000000000000", nil)

	req := model.CalculateMintRequest{
		Address:    "0x123",
		MintAmount: "10000000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateMint(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.NotEmpty(t, projection.NewDebt)
	assert.Equal(t, "100000000000000000000", projection.NewCollateralValue)
	mockCache.AssertExpectations(t)
}

func TestCalculateMint_NoCollateral(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("", assert.AnError)
	mockCache.On("HGet", "user:debt", "0x123").Return("0", nil)

	req := model.CalculateMintRequest{
		Address:    "0x123",
		MintAmount: "10000000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateMint(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.Equal(t, "10000000000000000000", projection.NewDebt)
}

func TestCalculateBurn_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x456").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x456").Return("50000000000000000000", nil)

	req := model.CalculateBurnRequest{
		Address:    "0x456",
		BurnAmount: "10000000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateBurn(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.Equal(t, "40000000000000000000", projection.NewDebt)
	assert.Equal(t, "100000000000000000000", projection.NewCollateralValue)
	mockCache.AssertExpectations(t)
}

func TestCalculateBurn_ExceedsDebt(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x789").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x789").Return("10000000000000000000", nil)

	req := model.CalculateBurnRequest{
		Address:    "0x789",
		BurnAmount: "20000000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateBurn(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.Equal(t, "0", projection.NewDebt)
}

func TestCalculateDeposit_Success(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	defer delete(constants.CollateralTokens, "ETH")

	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("50000000000000000000", nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)

	req := model.CalculateDepositRequest{
		Address:       "0x123",
		TokenAddress:  "0xethaddress",
		DepositAmount: "1000000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateDeposit(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.Equal(t, "50000000000000000000", projection.NewDebt)
	assert.NotEmpty(t, projection.NewCollateralValue)
	mockCache.AssertExpectations(t)
	mockPriceFeed.AssertExpectations(t)
}

func TestCalculateDeposit_PriceError(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	defer delete(constants.CollateralTokens, "ETH")

	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("50000000000000000000", nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("", assert.AnError)

	req := model.CalculateDepositRequest{
		Address:       "0x123",
		TokenAddress:  "0xethaddress",
		DepositAmount: "1000000000000000000",
	}

	ctx := context.Background()
	_, err := service.CalculateDeposit(ctx, req)

	assert.Error(t, err)
	mockPriceFeed.AssertExpectations(t)
}

func TestCalculateRedeem_Success(t *testing.T) {
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer delete(constants.CollateralTokens, "BTC")

	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewHealthFactorCalculationService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("50000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	req := model.CalculateRedeemRequest{
		Address:      "0x123",
		TokenAddress: "0xbtcaddress",
		RedeemAmount: "500000000000000000",
	}

	ctx := context.Background()
	projection, err := service.CalculateRedeem(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, projection.HealthFactorAfter)
	assert.Equal(t, "50000000000000000000", projection.NewDebt)
	assert.NotEmpty(t, projection.NewCollateralValue)
	mockCache.AssertExpectations(t)
	mockPriceFeed.AssertExpectations(t)
}

func TestGetTokenNameByAddress_ETH(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	defer delete(constants.CollateralTokens, "ETH")

	tokenName := getTokenNameByAddress("0xethaddress")
	assert.Equal(t, "ETH", tokenName)
}

func TestGetTokenNameByAddress_BTC(t *testing.T) {
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer delete(constants.CollateralTokens, "BTC")

	tokenName := getTokenNameByAddress("0xbtcaddress")
	assert.Equal(t, "BTC", tokenName)
}

func TestGetTokenNameByAddress_Unknown(t *testing.T) {
	tokenName := getTokenNameByAddress("0xunknown")
	assert.Equal(t, "", tokenName)
}

func TestGetTokenPrice_ETH(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)
	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)

	price, err := getTokenPrice(mockPriceFeed, "ETH")

	assert.NoError(t, err)
	assert.Equal(t, "3000000000000000000000", price)
	mockPriceFeed.AssertExpectations(t)
}

func TestGetTokenPrice_BTC(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	price, err := getTokenPrice(mockPriceFeed, "BTC")

	assert.NoError(t, err)
	assert.Equal(t, "50000000000000000000000", price)
	mockPriceFeed.AssertExpectations(t)
}

func TestGetTokenPrice_Unknown(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)

	price, err := getTokenPrice(mockPriceFeed, "UNKNOWN")

	assert.NoError(t, err)
	assert.Equal(t, "0", price)
}
