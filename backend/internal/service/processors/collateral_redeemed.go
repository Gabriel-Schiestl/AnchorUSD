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

func ProcessCollateralRedeemed(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	logger := utils.GetLogger()
	logger.Info().Str("event", eventName).Uint64("block", log.BlockNumber).Uint("index", log.Index).Msg("Processing collateral redeemed event")

	event := decodeEventData(log)
	if event == nil {
		logger.Error().Str("event", eventName).Uint64("block", log.BlockNumber).Msg("Failed to decode collateral redeemed event")
		return
	}

	logger.Debug().Str("user", event.From.Hex()).Str("token", event.Token.Hex()).Str("amount", event.Amount.String()).Msg("Event decoded successfully")
	
	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		logger.Error().Err(err).Str("event", eventName).Msg("Failed to create event in database")
		return
	}
	logger.Debug().Uint("event_id", eventModel.ID).Msg("Event record created in database")

	collateral := &model.Redeem{
		ID:                uuid.New().String(),
		EventID:           eventModel.ID,
		UserAddress:       event.From.Hex(),
		CollateralAddress: event.Token.Hex(),
		Amount:            model.NewBigInt(event.Amount),
	}

	err = storage.GetCollateralStore().CreateRedeem(context.Background(), collateral)
	if err != nil {
		logger.Error().Err(err).Str("user", event.From.Hex()).Msg("Failed to create redeem record")
		return
	}
	logger.Debug().Str("redeem_id", collateral.ID).Msg("Redeem record created")

	metric := model.Metrics{
		UserAddress: event.From,
		Amount:      event.Amount,
		Asset:       model.CollateralAsset,
		Operation:   model.Subtraction,
		BlockNumber: eventModel.BlockNumber,
	}

	metricsChan <- metric
	logger.Info().Str("user", event.From.Hex()).Str("token", event.Token.Hex()).Str("amount", event.Amount.String()).Msg("Collateral redeemed event processed and metric sent to channel")
}

func decodeEventData(log types.Log) *model.CollateralRedeemedEvent {
	event := &model.CollateralRedeemedEvent{}

	abi, err := blockchain.GetABI()
	if err != nil {
		return nil
	}

	err = abi.UnpackIntoInterface(event, "CollateralRedeemed", log.Data)
	if err != nil {
		return nil
	}

	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.To = common.HexToAddress(log.Topics[2].Hex())

	return event
}
