package service

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateLiquidations_Success(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)
	mockCache := new(MockCacheStore)

	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)

	btcCollateral := map[string]string{
		"0x123": "1000000000000000000", // 1 BTC
	}
	ethCollateral := map[string]string{
		"0x456": "2000000000000000000", // 2 ETH
	}

	mockCache.On("HGetAll", mock.MatchedBy(func(key string) bool {
		return key == "collateral:0xbtcaddress"
	})).Return(btcCollateral, nil)

	mockCache.On("HGetAll", mock.MatchedBy(func(key string) bool {
		return key == "collateral:0xethaddress"
	})).Return(ethCollateral, nil)

	mockCache.On("HGet", "user:debt", mock.Anything).Return("1000000000000000000", nil)
	mockCache.On("HSet", "user:collateral_usd", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("HSet", "user:health_factor", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("HSet", "liquidatable", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("HSet", "collateral", "total_supply", mock.Anything).Return(nil)

	err := CalculateLiquidations(mockPriceFeed, mockCache)

	assert.NoError(t, err)
	mockPriceFeed.AssertExpectations(t)
	mockCache.AssertCalled(t, "HSet", "collateral", "total_supply", mock.Anything)
}

func TestCalculateLiquidations_BTCPriceError(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)
	mockCache := new(MockCacheStore)

	mockPriceFeed.On("GetBtcUsdPrice").Return("", errors.New("price feed error"))

	err := CalculateLiquidations(mockPriceFeed, mockCache)

	assert.Error(t, err)
	assert.EqualError(t, err, "price feed error")
	mockPriceFeed.AssertExpectations(t)
}

func TestCalculateLiquidations_ETHPriceError(t *testing.T) {
	mockPriceFeed := new(MockPriceFeedAPI)
	mockCache := new(MockCacheStore)

	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("", errors.New("eth price error"))

	err := CalculateLiquidations(mockPriceFeed, mockCache)

	assert.Error(t, err)
	assert.EqualError(t, err, "eth price error")
	mockPriceFeed.AssertExpectations(t)
}

func TestUpdateBTCCollateralMetrics(t *testing.T) {
	collateralSupply := map[string]string{
		"0x123": "1000000000000000000",
		"0x456": "2000000000000000000",
	}
	btcPrice := "50000000000000000000000"
	totalUSDCollateral := make(map[string]*big.Int)

	updateBTCCollateralMetrics(collateralSupply, btcPrice, totalUSDCollateral)

	assert.Equal(t, 2, len(totalUSDCollateral))
	assert.NotNil(t, totalUSDCollateral["0x123"])
	assert.NotNil(t, totalUSDCollateral["0x456"])
}

func TestUpdateETHCollateralMetrics(t *testing.T) {
	collateralSupply := map[string]string{
		"0x789": "1500000000000000000",
	}
	ethPrice := "3000000000000000000000"
	totalUSDCollateral := make(map[string]*big.Int)

	updateETHCollateralMetrics(collateralSupply, ethPrice, totalUSDCollateral)

	assert.Equal(t, 1, len(totalUSDCollateral))
	assert.NotNil(t, totalUSDCollateral["0x789"])
}

func TestUpdateBTCCollateralMetrics_InvalidAmount(t *testing.T) {
	collateralSupply := map[string]string{
		"0x123": "invalid",
	}
	btcPrice := "50000000000000000000000"
	totalUSDCollateral := make(map[string]*big.Int)

	updateBTCCollateralMetrics(collateralSupply, btcPrice, totalUSDCollateral)

	assert.Equal(t, 0, len(totalUSDCollateral))
}

func TestUpdateLiquidationHealthFactors(t *testing.T) {
	mockCache := new(MockCacheStore)
	totalCollateralSupply := new(big.Int)
	totalUSDCollateralByUser := map[string]*big.Int{
		"0x123": big.NewInt(50000),
	}

	mockCache.On("HSet", "user:collateral_usd", "0x123", "50000").Return(nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("10000", nil)
	mockCache.On("HSet", "user:health_factor", "0x123", mock.Anything).Return(nil)

	updateLiquidationHealthFactors(totalCollateralSupply, mockCache, totalUSDCollateralByUser)

	assert.Equal(t, "50000", totalCollateralSupply.String())
	mockCache.AssertExpectations(t)
}
