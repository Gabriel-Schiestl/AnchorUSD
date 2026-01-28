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

func (s *collateralStore) GetLatestDeposits(ctx context.Context, userAddress string, limit int) ([]model.Deposit, error) {
	var deposits []model.Deposit
	err := s.DB.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Order("event_id DESC").
		Limit(limit).
		Find(&deposits).Error
	return deposits, err
}

func (s *collateralStore) GetLatestRedeems(ctx context.Context, userAddress string, limit int) ([]model.Redeem, error) {
	var redeems []model.Redeem
	err := s.DB.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Order("event_id DESC").
		Limit(limit).
		Find(&redeems).Error
	return redeems, err
}

func (s *collateralStore) IterateTotalDepositedGroupingByUser(ctx context.Context, limit int, cb func(map[string]map[string]*big.Int) error) error {
	if limit <= 0 {
        limit = 500
    }

    var count int64
    if err := s.DB.WithContext(ctx).Model(&model.Deposit{}).Distinct("user_address").Count(&count).Error; err != nil {
        return err
    }
    pages := int((count + int64(limit) - 1) / int64(limit))

	for page := range pages {
		rows, err := s.DB.WithContext(ctx).
            Raw("SELECT user_address, SUM(amount)::text, collateral_address FROM deposits GROUP BY user_address, collateral_address ORDER BY user_address LIMIT ? OFFSET ?", limit, page*limit).
            Rows()
        if err != nil {
            return err
        }

		mappings := make(map[string]map[string]*big.Int)

        for rows.Next() {
            var user string
            var totalStr string
			var collateralAddr string
            if err := rows.Scan(&user, &totalStr, &collateralAddr); err != nil {
                rows.Close()
                return err
            }
            bi := new(big.Int)
            bi.SetString(totalStr, 10)
			if _, exists := mappings[user]; !exists {
				mappings[user] = make(map[string]*big.Int)
			}
			mappings[user][collateralAddr] = bi
        }
		if err := cb(mappings); err != nil {
			rows.Close()
			return err
		}
        rows.Close()
	}

	return nil
}

func (s *collateralStore) GetTotalCollateralRedeemedGroupingByUser(ctx context.Context, users []string) (map[string]map[string]*big.Int, error) {
	var results []struct {
		UserAddress      string
		CollateralAddress string
		TotalRedeemed    model.BigInt
	}

	err := s.DB.WithContext(ctx).
		Model(&model.Redeem{}).
		Select("user_address, collateral_address, SUM(amount) as total_redeemed").
		Group("user_address, collateral_address").
		Where("user_address IN ?", users).
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
