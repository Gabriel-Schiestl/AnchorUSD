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

func (cs *coinStore) GetLatestBurns(ctx context.Context, userAddress string, limit int) ([]model.Burns, error) {
	var burns []model.Burns
	err := cs.DB.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Order("event_id DESC").
		Limit(limit).
		Find(&burns).Error
	return burns, err
}

func (cs *coinStore) GetTotalBurnedGroupingByUser(ctx context.Context, users []string) (map[string]*big.Int, error) {
	var results []struct {
		UserAddress string
		TotalBurned model.BigInt
	}

	err := cs.DB.WithContext(ctx).
		Model(&model.Burns{}).
		Select("user_address, SUM(amount) as total_burned").
		Group("user_address").
		Where("user_address IN ?", users).
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

func (cs *coinStore) GetLatestMints(ctx context.Context, userAddress string, limit int) ([]model.Mints, error) {
	var mints []model.Mints
	err := cs.DB.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Order("event_id DESC").
		Limit(limit).
		Find(&mints).Error
	return mints, err
}

func (cs *coinStore) IterateTotalMintedGroupingByUser(ctx context.Context, limit int, cb func(map[string]*big.Int) error) error {
    if limit <= 0 {
        limit = 500
    }

    var count int64
    if err := cs.DB.WithContext(ctx).Model(&model.Mints{}).Distinct("user_address").Count(&count).Error; err != nil {
        return err
    }
    pages := int((count + int64(limit) - 1) / int64(limit))

    for page := range pages {
        rows, err := cs.DB.WithContext(ctx).
            Raw("SELECT user_address, SUM(amount)::text as total_minted FROM mints GROUP BY user_address ORDER BY user_address LIMIT ? OFFSET ?", limit, page*limit).
            Rows()
        if err != nil {
            return err
        }

		mappings := make(map[string]*big.Int)

        for rows.Next() {
            var user string
            var totalStr string
            if err := rows.Scan(&user, &totalStr); err != nil {
                rows.Close()
                return err
            }
            bi := new(big.Int)
            bi.SetString(totalStr, 10)
			mappings[user] = bi
        }
		if err := cb(mappings); err != nil {
			rows.Close()
			return err
		}
        rows.Close()
    }

    return nil
}
