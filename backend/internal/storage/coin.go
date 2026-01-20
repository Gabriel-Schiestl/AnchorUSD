package storage

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
)

var coinStr coinStore

type coinStore struct {
	DB *gorm.DB
}

func NewCoinStore(db *gorm.DB) *coinStore {
	coinStr = coinStore{DB: db}
	return &coinStr
}

func GetCoinStore() *coinStore {
	return &coinStr
}

func (cs *coinStore) CreateBurn(ctx context.Context, burn *model.Burns) error {
	return cs.DB.WithContext(ctx).Create(burn).Error
}

// TODO: implement pagination for large datasets
func (cs *coinStore) GetTotalBurnedGroupingByUser(ctx context.Context) (map[string]*big.Int, error) {
	var results []struct {
		UserAddress string
		TotalBurned model.BigInt
	}

	err := cs.DB.WithContext(ctx).
		Model(&model.Burns{}).
		Select("user_address, SUM(amount) as total_burned").
		Group("user_address").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	totalBurnedByUser := make(map[string]*big.Int)
	for _, result := range results {
		totalBurnedByUser[result.UserAddress] = result.TotalBurned.Int
	}

	return totalBurnedByUser, nil
}

func (cs *coinStore) CreateMint(ctx context.Context, mint *model.Mints) error {
	return cs.DB.WithContext(ctx).Create(mint).Error
}

// TODO: implement pagination for large datasets
func (cs *coinStore) GetTotalMintedGroupingByUser(ctx context.Context) (map[string]*big.Int, error) {
	var results []struct {
		UserAddress string
		TotalMinted model.BigInt
	}

	err := cs.DB.WithContext(ctx).
		Model(&model.Mints{}).
		Select("user_address, SUM(amount) as total_minted").
		Group("user_address").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	totalMintedByUser := make(map[string]*big.Int)
	for _, result := range results {
		totalMintedByUser[result.UserAddress] = result.TotalMinted.Int
	}

	return totalMintedByUser, nil
}
