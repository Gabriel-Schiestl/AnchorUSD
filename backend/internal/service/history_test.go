package service

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCollateralHistoryStore struct {
	mock.Mock
}

func (m *MockCollateralHistoryStore) GetLatestDeposits(ctx context.Context, userAddress string, limit int) ([]model.Deposit, error) {
	args := m.Called(ctx, userAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Deposit), args.Error(1)
}

func (m *MockCollateralHistoryStore) GetLatestRedeems(ctx context.Context, userAddress string, limit int) ([]model.Redeem, error) {
	args := m.Called(ctx, userAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Redeem), args.Error(1)
}

type MockCoinHistoryStore struct {
	mock.Mock
}

func (m *MockCoinHistoryStore) GetLatestMints(ctx context.Context, userAddress string, limit int) ([]model.Mints, error) {
	args := m.Called(ctx, userAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Mints), args.Error(1)
}

func (m *MockCoinHistoryStore) GetLatestBurns(ctx context.Context, userAddress string, limit int) ([]model.Burns, error) {
	args := m.Called(ctx, userAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Burns), args.Error(1)
}

type MockLiquidationHistoryStore struct {
	mock.Mock
}

func (m *MockLiquidationHistoryStore) GetLatestLiquidations(ctx context.Context, userAddress string, limit int) ([]model.Liquidations, error) {
	args := m.Called(ctx, userAddress, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Liquidations), args.Error(1)
}

type MockEventHistoryStore struct {
	mock.Mock
}

func (m *MockEventHistoryStore) GetEventByID(ctx context.Context, eventID uint) (*model.Events, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Events), args.Error(1)
}

func TestNewHistoryService(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	assert.NotNil(t, service)
	assert.Equal(t, mockCollateral, service.collateralStore)
	assert.Equal(t, mockCoin, service.coinStore)
	assert.Equal(t, mockLiquidation, service.liquidationStore)
	assert.Equal(t, mockEvents, service.eventsStore)
}

func TestGetUserHistory_Success(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	deposits := []model.Deposit{
		{
			ID:                "1",
			EventID:           100,
			CollateralAddress: "0xeth",
			Amount:            model.BigInt{Int: big.NewInt(1000)},
		},
	}
	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return(deposits, nil)

	redeems := []model.Redeem{
		{
			ID:                "2",
			EventID:           101,
			CollateralAddress: "0xbtc",
			Amount:            model.BigInt{Int: big.NewInt(500)},
		},
	}
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return(redeems, nil)

	mints := []model.Mints{
		{
			ID:      "3",
			EventID: 102,
			Amount:  model.BigInt{Int: big.NewInt(2000)},
		},
	}
	mockCoin.On("GetLatestMints", ctx, userAddress, 10).Return(mints, nil)

	burns := []model.Burns{
		{
			ID:      "4",
			EventID: 103,
			Amount:  model.BigInt{Int: big.NewInt(1000)},
		},
	}
	mockCoin.On("GetLatestBurns", ctx, userAddress, 10).Return(burns, nil)

	liquidations := []model.Liquidations{
		{
			ID:                "5",
			EventID:           104,
			CollateralAddress: "0xeth",
			DebtCovered:       model.BigInt{Int: big.NewInt(3000)},
		},
	}
	mockLiquidation.On("GetLatestLiquidations", ctx, userAddress, 10).Return(liquidations, nil)

	event := &model.Events{
		ID:        100,
		TxHash:    "0xabcd",
		CreatedAt: 1609459200,
	}
	mockEvents.On("GetEventByID", ctx, mock.AnythingOfType("uint")).Return(event, nil)

	history, err := service.GetUserHistory(ctx, userAddress)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(history.Deposits), 0)
	assert.GreaterOrEqual(t, len(history.MintBurn), 0)
	assert.GreaterOrEqual(t, len(history.Liquidations), 0)
	mockCollateral.AssertExpectations(t)
	mockCoin.AssertExpectations(t)
	mockLiquidation.AssertExpectations(t)
}

func TestGetUserHistory_DepositsError(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return(nil, errors.New("database error"))

	_, err := service.GetUserHistory(ctx, userAddress)

	assert.Error(t, err)
	assert.EqualError(t, err, "database error")
}

func TestGetUserHistory_RedeemsError(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return([]model.Deposit{}, nil)
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return(nil, errors.New("redeems error"))

	_, err := service.GetUserHistory(ctx, userAddress)

	assert.Error(t, err)
	assert.EqualError(t, err, "redeems error")
}

func TestGetUserHistory_MintsError(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return([]model.Deposit{}, nil)
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return([]model.Redeem{}, nil)
	mockCoin.On("GetLatestMints", ctx, userAddress, 10).Return(nil, errors.New("mints error"))

	_, err := service.GetUserHistory(ctx, userAddress)

	assert.Error(t, err)
	assert.EqualError(t, err, "mints error")
}

func TestGetUserHistory_BurnsError(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return([]model.Deposit{}, nil)
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return([]model.Redeem{}, nil)
	mockCoin.On("GetLatestMints", ctx, userAddress, 10).Return([]model.Mints{}, nil)
	mockCoin.On("GetLatestBurns", ctx, userAddress, 10).Return(nil, errors.New("burns error"))

	_, err := service.GetUserHistory(ctx, userAddress)

	assert.Error(t, err)
	assert.EqualError(t, err, "burns error")
}

func TestGetUserHistory_LiquidationsError(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x123"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return([]model.Deposit{}, nil)
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return([]model.Redeem{}, nil)
	mockCoin.On("GetLatestMints", ctx, userAddress, 10).Return([]model.Mints{}, nil)
	mockCoin.On("GetLatestBurns", ctx, userAddress, 10).Return([]model.Burns{}, nil)
	mockLiquidation.On("GetLatestLiquidations", ctx, userAddress, 10).Return(nil, errors.New("liquidations error"))

	_, err := service.GetUserHistory(ctx, userAddress)

	assert.Error(t, err)
	assert.EqualError(t, err, "liquidations error")
}

func TestGetUserHistory_EmptyResults(t *testing.T) {
	mockCollateral := new(MockCollateralHistoryStore)
	mockCoin := new(MockCoinHistoryStore)
	mockLiquidation := new(MockLiquidationHistoryStore)
	mockEvents := new(MockEventHistoryStore)

	service := NewHistoryService(mockCollateral, mockCoin, mockLiquidation, mockEvents)

	ctx := context.Background()
	userAddress := "0x999"

	mockCollateral.On("GetLatestDeposits", ctx, userAddress, 10).Return([]model.Deposit{}, nil)
	mockCollateral.On("GetLatestRedeems", ctx, userAddress, 10).Return([]model.Redeem{}, nil)
	mockCoin.On("GetLatestMints", ctx, userAddress, 10).Return([]model.Mints{}, nil)
	mockCoin.On("GetLatestBurns", ctx, userAddress, 10).Return([]model.Burns{}, nil)
	mockLiquidation.On("GetLatestLiquidations", ctx, userAddress, 10).Return([]model.Liquidations{}, nil)

	history, err := service.GetUserHistory(ctx, userAddress)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(history.Deposits))
	assert.Equal(t, 0, len(history.MintBurn))
	assert.Equal(t, 0, len(history.Liquidations))
}
