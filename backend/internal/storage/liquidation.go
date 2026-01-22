package storage

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"gorm.io/gorm"
)

var liquidationStr liquidationStore

type liquidationStore struct {
	DB *gorm.DB
}

func NewLiquidationStore(db *gorm.DB) *liquidationStore {
	liquidationStr = liquidationStore{DB: db}
	return &liquidationStr
}

func GetLiquidationStore() *liquidationStore {
	return &liquidationStr
}

func (ls *liquidationStore) CreateLiquidation(ctx context.Context, liquidation *model.Liquidations) error {
	logger := utils.GetLogger()
	logger.Debug().Str("liquidated_user", liquidation.LiquidatedUserAddress).Str("liquidator", liquidation.LiquidatorAddress).Msg("Creating liquidation record")

	err := ls.DB.WithContext(ctx).Create(liquidation).Error
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create liquidation record in database")
		return err
	}

	logger.Info().Str("liquidation_id", liquidation.ID).Msg("Liquidation record created successfully")
	return nil
}
