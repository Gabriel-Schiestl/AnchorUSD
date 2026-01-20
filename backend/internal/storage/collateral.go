package storage

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
)

var collatStore collateralStore

type collateralStore struct {
	DB *gorm.DB
}

func NewCollateralStore(db *gorm.DB) *collateralStore {
	collatStore = collateralStore{DB: db}
	return &collatStore
}

func GetCollateralStore() *collateralStore {
	return &collatStore
}

func (s *collateralStore) CreateRedeem(ctx context.Context, redeem *model.Redeem) error {
	result := s.DB.WithContext(ctx).Create(redeem)
	return result.Error
}

func (s *collateralStore) CreateDeposit(ctx context.Context, deposit *model.Deposit) error {
	result := s.DB.WithContext(ctx).Create(deposit)
	return result.Error
}

func (s *collateralStore) GetTotalCollateralDepositedGroupingByUser(ctx context.Context) (map[string]map[string]*big.Int, error) {
	var results []struct {
		UserAddress    string
		CollateralAddress string
		TotalDeposited model.BigInt
	}

	err := s.DB.WithContext(ctx).
		Model(&model.Deposit{}).
		Select("user_address, collateral_address, SUM(amount) as total_deposited").
		Group("user_address, collateral_address").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	totalCollateralByUser := make(map[string]map[string]*big.Int)
	for _, result := range results {
		if _, exists := totalCollateralByUser[result.UserAddress]; !exists {
			totalCollateralByUser[result.UserAddress] = make(map[string]*big.Int)
		}
		totalCollateralByUser[result.UserAddress][result.CollateralAddress] = result.TotalDeposited.Int
	}

	return totalCollateralByUser, nil
}

func (s *collateralStore) GetTotalCollateralRedeemedGroupingByUser(ctx context.Context) (map[string]map[string]*big.Int, error) {
	var results []struct {
		UserAddress      string
		CollateralAddress string
		TotalRedeemed    model.BigInt
	}

	err := s.DB.WithContext(ctx).
		Model(&model.Redeem{}).
		Select("user_address, collateral_address, SUM(amount) as total_redeemed").
		Group("user_address, collateral_address").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	totalRedeemedByUser := make(map[string]map[string]*big.Int)
	for _, result := range results {
		if _, exists := totalRedeemedByUser[result.UserAddress]; !exists {
			totalRedeemedByUser[result.UserAddress] = make(map[string]*big.Int)
		}
		totalRedeemedByUser[result.UserAddress][result.CollateralAddress] = result.TotalRedeemed.Int
	}

	return totalRedeemedByUser, nil
}
