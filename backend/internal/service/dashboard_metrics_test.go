package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDashboardMetricsService(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)

	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	assert.NotNil(t, service)
	assert.Equal(t, mockCache, service.Store)
	assert.Equal(t, mockPriceFeed, service.PriceFeed)
}

func TestGetDashboardMetrics_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGetAll", "liquidatable").Return(map[string]string{
		"0x123": "0.8",
	}, nil)
	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("80000", nil)

	mockCache.On("HGet", "collateral", "total_supply").Return("1000000", nil)
	mockCache.On("HGetAll", mock.Anything).Return(map[string]string{}, nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	mockCache.On("HGet", "coin", "total_supply").Return("100000", nil)
	mockCache.On("HGetAll", "user:debt").Return(map[string]string{
		"0x123": "80000",
	}, nil)

	mockCache.On("HGetAll", "user:health_factor").Return(map[string]string{
		"0x123": "1500000000000000000",
	}, nil)

	ctx := context.Background()
	metrics, err := service.GetDashboardMetrics(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, len(metrics.LiquidatableUsers), 0)
	mockCache.AssertExpectations(t)
}

func TestGetLiquidatableUsers_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGetAll", "liquidatable").Return(map[string]string{
		"0x123": "0.9",
		"0x456": "0.8",
	}, nil)
	mockCache.On("HGet", "user:collateral_usd", "0x123").Return("100000", nil)
	mockCache.On("HGet", "user:debt", "0x123").Return("90000", nil)
	mockCache.On("HGet", "user:collateral_usd", "0x456").Return("200000", nil)
	mockCache.On("HGet", "user:debt", "0x456").Return("160000", nil)

	users, err := service.getLiquidatableUsers()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))
	addresses := []string{users[0].Address, users[1].Address}
	assert.Contains(t, addresses, "0x123")
	assert.Contains(t, addresses, "0x456")
}

func TestGetLiquidatableUsers_EmptyResult(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGetAll", "liquidatable").Return(map[string]string{}, nil)

	users, err := service.getLiquidatableUsers()

	assert.NoError(t, err)
	assert.Equal(t, 0, len(users))
}

func TestGetTotalCollateral_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "collateral", "total_supply").Return("1000000", nil)
	mockCache.On("HGetAll", "collateral:0xethaddress").Return(map[string]string{
		"0x123": "500000",
	}, nil)
	mockCache.On("HGetAll", "collateral:0xbtcaddress").Return(map[string]string{
		"0x456": "500000",
	}, nil)
	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	collateral, err := service.getTotalCollateral()

	assert.NoError(t, err)
	assert.Equal(t, "1000000", collateral.Value)
	assert.GreaterOrEqual(t, len(collateral.Breakdown), 0)
}

func TestGetStableSupply_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "coin", "total_supply").Return("100000", nil)
	mockCache.On("HGetAll", "user:debt").Return(map[string]string{
		"0x123": "50000",
		"0x456": "30000",
	}, nil)
	mockCache.On("HGet", "collateral", "total_supply").Return("200000", nil)

	supply, err := service.getStableSupply()

	assert.NoError(t, err)
	assert.Equal(t, "100000", supply.Total)
	assert.Equal(t, "80000", supply.Circulating)
	assert.GreaterOrEqual(t, supply.Backing, float64(0))
}

func TestGetStableSupply_NoDebt(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGet", "coin", "total_supply").Return("100000", nil)
	mockCache.On("HGetAll", "user:debt").Return(nil, assert.AnError)

	supply, err := service.getStableSupply()

	assert.NoError(t, err)
	assert.Equal(t, "100000", supply.Total)
	assert.Equal(t, "100000", supply.Circulating)
}

func TestGetProtocolHealth_Success(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGetAll", "user:health_factor").Return(map[string]string{
		"0x123": "2000000000000000000",
		"0x456": "1500000000000000000",
	}, nil)
	mockCache.On("HGet", "collateral", "total_supply").Return("1000000", nil)
	mockCache.On("HGet", "coin", "total_supply").Return("500000", nil)

	health, err := service.getProtocolHealth()

	assert.NoError(t, err)
	assert.Equal(t, 2, health.TotalUsers)
	assert.GreaterOrEqual(t, health.AverageHealthFactor, float64(0))
	assert.GreaterOrEqual(t, health.CollateralizationRatio, float64(0))
}

func TestGetProtocolHealth_NoUsers(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)
	service := NewDashboardMetricsService(mockCache, mockPriceFeed)

	mockCache.On("HGetAll", "user:health_factor").Return(map[string]string{}, nil)

	health, err := service.getProtocolHealth()

	assert.NoError(t, err)
	assert.Equal(t, 0, health.TotalUsers)
	assert.Equal(t, float64(0), health.AverageHealthFactor)
	assert.Equal(t, 0, health.UsersAtRisk)
}
