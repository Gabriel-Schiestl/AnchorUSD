package storage

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
)

var collatStore collateralStore

type collateralStore struct {
	DB *gorm.DB
}

func NewCollateralStore(db *gorm.DB) *collateralStore {
	return &collateralStore{
		DB: db,
	}
}

func GetCollateralStore() *collateralStore {
	return &collatStore
}

func (s *collateralStore) Get(ctx context.Context) (*model.Events, error) {
	// Implementation to retrieve events from the database
	return &model.Events{}, nil
}

func (s *collateralStore) CreateRedeem(ctx context.Context, redeem *model.Redeem) error {
	result := s.DB.WithContext(ctx).Create(redeem)
	return result.Error
}

func (s *collateralStore) CreateDeposit(ctx context.Context, deposit *model.Deposit) error {
	result := s.DB.WithContext(ctx).Create(deposit)
	return result.Error
}

func (s *collateralStore) CreateMint(ctx context.Context, mint *model.Mints) error {
	result := s.DB.WithContext(ctx).Create(mint)
	return result.Error
}

func (s *collateralStore) CreateBurn(ctx context.Context, burn *model.Burns) error {
	result := s.DB.WithContext(ctx).Create(burn)
	return result.Error
}
