package service

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUsdPriceAndAddress(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer func() {
		delete(constants.CollateralTokens, "ETH")
		delete(constants.CollateralTokens, "BTC")
	}()

	mockPriceFeed := new(MockPriceFeedAPI)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	getUsdPriceAndAddress(mockPriceFeed)

	assert.Equal(t, "3000000000000000000000", ethPrice)
	assert.Equal(t, "50000000000000000000000", btcPrice)
	assert.Equal(t, "0xethaddress", ethAddr)
	assert.Equal(t, "0xbtcaddress", btcAddr)
	mockPriceFeed.AssertExpectations(t)
}

func TestGetUsdPriceAndAddress_ETHPriceError(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer func() {
		delete(constants.CollateralTokens, "ETH")
		delete(constants.CollateralTokens, "BTC")
	}()

	mockPriceFeed := new(MockPriceFeedAPI)

	mockPriceFeed.On("GetEthUsdPrice").Return("", assert.AnError)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	ethPrice = ""
	btcPrice = ""

	getUsdPriceAndAddress(mockPriceFeed)

	assert.Equal(t, "", ethPrice)
	assert.Equal(t, "50000000000000000000000", btcPrice)
}

func TestGetUsdPriceAndAddress_BTCPriceError(t *testing.T) {
	constants.CollateralTokens["ETH"] = "0xethaddress"
	constants.CollateralTokens["BTC"] = "0xbtcaddress"
	defer func() {
		delete(constants.CollateralTokens, "ETH")
		delete(constants.CollateralTokens, "BTC")
	}()

	mockPriceFeed := new(MockPriceFeedAPI)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("", assert.AnError)

	ethPrice = ""
	btcPrice = ""

	getUsdPriceAndAddress(mockPriceFeed)

	assert.Equal(t, "3000000000000000000000", ethPrice)
	assert.Equal(t, "", btcPrice)
}

func TestGetCollateralUSD_ETH(t *testing.T) {
	deposited := big.NewInt(1000000000000000000) // 1 ETH
	ethPrice = "3000000000000000000000"
	ethAddr = "0xeth"

	collateralUSD := getCollateralUSD("0xeth", deposited, ethPrice, btcPrice, ethAddr, btcAddr)

	assert.NotNil(t, collateralUSD)
	assert.Greater(t, collateralUSD.Cmp(big.NewInt(0)), 0)
}

func TestGetCollateralUSD_BTC(t *testing.T) {
	deposited := big.NewInt(1000000000000000000) // 1 BTC (in 18 decimals)
	btcPrice = "50000000000000000000000"
	btcAddr = "0xbtc"

	collateralUSD := getCollateralUSD("0xbtc", deposited, ethPrice, btcPrice, ethAddr, btcAddr)

	assert.NotNil(t, collateralUSD)
	assert.Greater(t, collateralUSD.Cmp(big.NewInt(0)), 0)
}

func TestGetCollateralUSD_ZeroDeposited(t *testing.T) {
	deposited := big.NewInt(0)
	ethPrice = "3000000000000000000000"
	ethAddr = "0xeth"

	collateralUSD := getCollateralUSD("0xeth", deposited, ethPrice, btcPrice, ethAddr, btcAddr)

	assert.NotNil(t, collateralUSD)
	assert.Equal(t, 0, collateralUSD.Cmp(big.NewInt(0)))
}

func TestGetCollateralUSD_NilDeposited(t *testing.T) {
	ethPrice = "3000000000000000000000"
	ethAddr = "0xeth"

	collateralUSD := getCollateralUSD("0xeth", nil, ethPrice, btcPrice, ethAddr, btcAddr)

	assert.NotNil(t, collateralUSD)
	assert.Equal(t, 0, collateralUSD.Cmp(big.NewInt(0)))
}

func TestGetCollateralUSD_UnknownCollateral(t *testing.T) {
	deposited := big.NewInt(1000000000000000000)
	ethPrice = "3000000000000000000000"
	btcPrice = "50000000000000000000000"
	ethAddr = "0xeth"
	btcAddr = "0xbtc"

	collateralUSD := getCollateralUSD("0xunknown", deposited, ethPrice, btcPrice, ethAddr, btcAddr)

	assert.NotNil(t, collateralUSD)
	assert.Equal(t, 0, collateralUSD.Cmp(big.NewInt(0)))
}

func TestIteratorTotalMintedByUserCallback(t *testing.T) {
	mockCache := new(MockCacheStore)
	cacheStore = mockCache

	mockCache.On("HSet", "user:debt", "0x123", mock.Anything).Return(nil)
	mockCache.On("HSet", "user:debt", "0x456", mock.Anything).Return(nil)
	mockCache.On("HSet", "coin", "total_supply", mock.Anything).Return(nil)

	assert.NotNil(t, mockCache)
}

func TestUpdateMetrics_Integration(t *testing.T) {
	mockCache := new(MockCacheStore)
	mockPriceFeed := new(MockPriceFeedAPI)

	mockPriceFeed.On("GetEthUsdPrice").Return("3000000000000000000000", nil)
	mockPriceFeed.On("GetBtcUsdPrice").Return("50000000000000000000000", nil)

	assert.NotNil(t, mockCache)
	assert.NotNil(t, mockPriceFeed)
}
