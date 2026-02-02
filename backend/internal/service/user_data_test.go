package service

import (
	"context"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewUserDataService(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)

	service := NewUserDataService(mockCache, mockPriceFeed)

	assert.NotNil(t, service)
	assert.Equal(t, mockCache, service.Store)
	assert.Equal(t, mockPriceFeed, service.PriceFeed)
}

func TestGetUserData_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	ctx := context.Background()
	userAddress := "0x123"

	mockCache.On("HGet", "user:debt", userAddress).Return("50000000000000000000", nil)
	mockCache.On("HGet", "user:collateral_usd", userAddress).Return("200000000000000000000", nil)
	mockCache.On("HGet", "user:health_factor", userAddress).Return("2000000000000000000", nil)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", mock.Anything, userAddress).Return("", assert.AnError)

	userData, err := service.GetUserData(ctx, userAddress)

	assert.NoError(t, err)
	assert.Equal(t, "50000000000000000000", userData.TotalDebt)
	assert.Equal(t, "200000000000000000000", userData.CollateralValueUSD)
	assert.Equal(t, "2000000000000000000", userData.CurrentHealthFactor)
	assert.NotEmpty(t, userData.MaxMintable)
	assert.GreaterOrEqual(t, len(userData.CollateralDeposited), 0)
	mockCache.AssertExpectations(t)
}

func TestGetUserData_NoDebt(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	ctx := context.Background()
	userAddress := "0x456"

	mockCache.On("HGet", "user:debt", userAddress).Return("", assert.AnError)
	mockCache.On("HGet", "user:collateral_usd", userAddress).Return("100000000000000000000", nil)
	mockCache.On("HGet", "user:health_factor", userAddress).Return("0", nil)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xethaddress", userAddress).Return("", assert.AnError)
	mockCache.On("HGet", "collateral:0xbtcaddress", userAddress).Return("", assert.AnError)

	userData, err := service.GetUserData(ctx, userAddress)

	assert.NoError(t, err)
	assert.Equal(t, "0", userData.TotalDebt)
	assert.Equal(t, "100000000000000000000", userData.CollateralValueUSD)
}

func TestGetUserData_NoCollateral(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	ctx := context.Background()
	userAddress := "0x789"

	mockCache.On("HGet", "user:debt", userAddress).Return("10000000000000000000", nil)
	mockCache.On("HGet", "user:collateral_usd", userAddress).Return("", assert.AnError)
	mockCache.On("HGet", "user:health_factor", userAddress).Return("", assert.AnError)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xethaddress", userAddress).Return("", assert.AnError)
	mockCache.On("HGet", "collateral:0xbtcaddress", userAddress).Return("", assert.AnError)

	userData, err := service.GetUserData(ctx, userAddress)

	assert.NoError(t, err)
	assert.Equal(t, "10000000000000000000", userData.TotalDebt)
	assert.Equal(t, "0", userData.CollateralValueUSD)
	assert.Equal(t, "0", userData.CurrentHealthFactor)
}

func TestFetchCollateralAssets_MultipleAssets(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer func() {
		delete(constants.CollateralTokens, "ETH")
		delete(constants.CollateralTokens, "BTC")
	}()

	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	userAddress := "0x123"

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xethaddress", userAddress).Return("2000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xbtcaddress", userAddress).Return("1000000000000000000", nil)

	assets := service.fetchCollateralAssets(userAddress)

	assert.Equal(t, 2, len(assets))
}

func TestFetchCollateralAssets_InvalidAmount(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	userAddress := "0x456"

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xethaddress", userAddress).Return("invalid_number", nil)
	mockCache.On("HGet", "collateral:0xbtcaddress", userAddress).Return("", assert.AnError)

	assets := service.fetchCollateralAssets(userAddress)

	assert.Equal(t, 0, len(assets))
}

func TestFetchCollateralAssets_NoAssets(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewUserDataService(mockCache, mockPriceFeed)

	userAddress := "0x999"

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockCache.On("HGet", "collateral:0xethaddress", userAddress).Return("", assert.AnError)
	mockCache.On("HGet", "collateral:0xbtcaddress", userAddress).Return("", assert.AnError)

	assets := service.fetchCollateralAssets(userAddress)

	assert.Equal(t, 0, len(assets))
}
