package processors

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/blockchain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func ProcessLiquidation(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	logger := utils.GetLogger()
	logger.Info().Uint64("block", log.BlockNumber).Uint("index", log.Index).Msg("Processing liquidation event")

	event := decodeLiquidationEvent(log)
	if event == nil {
		logger.Error().Msg("Failed to decode liquidation event")
		return
	}

	logger.Debug().
		Str("liquidated_user", event.LiquidatedUser.Hex()).
		Str("liquidator", event.Liquidator.Hex()).
		Str("token", event.TokenCollateral.Hex()).
		Str("collateral_amount", event.CollateralAmount.String()).
		Str("debt_covered", event.DebtCovered.String()).
		Msg("Liquidation event decoded")

	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create event in database")
		return
	}

	liquidation := &model.Liquidations{
		ID:                    uuid.New().String(),
		EventID:               eventModel.ID,
		LiquidatedUserAddress: event.LiquidatedUser.Hex(),
		LiquidatorAddress:     event.Liquidator.Hex(),
		CollateralAddress:     event.TokenCollateral.Hex(),
		CollateralAmount:      model.NewBigInt(event.CollateralAmount),
		DebtCovered:           model.NewBigInt(event.DebtCovered),
	}

	err = storage.GetLiquidationStore().CreateLiquidation(context.Background(), liquidation)
	if err != nil {
		logger.Error().Err(err).Str("liquidation_id", liquidation.ID).Msg("Failed to create liquidation in database")
		return
	}

	logger.Info().
		Str("liquidated_user", event.LiquidatedUser.Hex()).
		Str("debt_covered", event.DebtCovered.String()).
		Msg("Liquidation event saved, sending metrics")

	metricsChan <- model.Metrics{
		UserAddress:            event.LiquidatedUser,
		Amount:                 event.CollateralAmount,
		Asset:                  model.CollateralAsset,
		Operation:              model.Subtraction,
		BlockNumber:            eventModel.BlockNumber,
		CollateralTokenAddress: event.TokenCollateral,
	}

	metricsChan <- model.Metrics{
		UserAddress: event.LiquidatedUser,
		Amount:      event.DebtCovered,
		Asset:       model.StablecoinAsset,
		Operation:   model.Subtraction,
		BlockNumber: eventModel.BlockNumber,
	}

	logger.Info().
		Str("liquidated_user", event.LiquidatedUser.Hex()).
		Str("liquidator", event.Liquidator.Hex()).
		Msg("Liquidation processed successfully and metrics sent")
}

func decodeLiquidationEvent(log types.Log) *model.LiquidationEvent {
	logger := utils.GetLogger()

	event := &model.LiquidationEvent{}

	abi, err := blockchain.GetABI()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get ABI")
		return nil
	}

	err = abi.UnpackIntoInterface(event, "Liquidation", log.Data)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to unpack liquidation event data")
		return nil
	}

	if len(log.Topics) < 4 {
		logger.Error().Int("topics_count", len(log.Topics)).Msg("Insufficient topics in liquidation event")
		return nil
	}

	event.LiquidatedUser = common.HexToAddress(log.Topics[1].Hex())
	event.Liquidator = common.HexToAddress(log.Topics[2].Hex())
	event.TokenCollateral = common.HexToAddress(log.Topics[3].Hex())

	logger.Debug().
		Str("liquidated_user", event.LiquidatedUser.Hex()).
		Str("liquidator", event.Liquidator.Hex()).
		Str("token", event.TokenCollateral.Hex()).
		Msg("Liquidation event decoded successfully")

	return event
}
