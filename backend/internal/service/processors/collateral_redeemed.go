package processors

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/metrics"
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
		metrics.RecordError("redeem", "decode_error")
		return
	}

	logger.Debug().Str("user", event.User.Hex()).Str("token", event.Token.Hex()).Str("amount", event.Amount.String()).Msg("Event decoded successfully")
	
	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		logger.Error().Err(err).Str("event", eventName).Msg("Failed to create event in database")
		metrics.RecordError("redeem", "database_error")
		return
	}
	logger.Debug().Uint("event_id", eventModel.ID).Msg("Event record created in database")

	collateral := &model.Redeem{
		ID:                uuid.New().String(),
		EventID:           eventModel.ID,
		UserAddress:       event.User.Hex(),
		CollateralAddress: event.Token.Hex(),
		Amount:            model.NewBigInt(event.Amount),
	}

	err = storage.GetCollateralStore().CreateRedeem(context.Background(), collateral)
	if err != nil {
		logger.Error().Err(err).Str("user", event.User.Hex()).Msg("Failed to create redeem record")
		metrics.RecordError("redeem", "database_error")
		return
	}
	logger.Debug().Str("redeem_id", collateral.ID).Msg("Redeem record created")

	tokenName := getTokenNameByAddress(event.Token.Hex())
	if tokenName != "" {
		metrics.CollateralRedeemsTotal.WithLabelValues(tokenName).Inc()
	}

	metric := model.Metrics{
		UserAddress: event.User,
		Amount:      event.Amount,
		Asset:       model.CollateralAsset,
		Operation:   model.Subtraction,
		BlockNumber: eventModel.BlockNumber,
	}

	metricsChan <- metric
	logger.Info().Str("user", event.User.Hex()).Str("token", event.Token.Hex()).Str("amount", event.Amount.String()).Msg("Collateral redeemed event processed and metric sent to channel")
}

func decodeEventData(log types.Log) *model.CollateralRedeemedEvent {
	if len(log.Topics) < 4 {
		return nil
	}

	event := &model.CollateralRedeemedEvent{}
	event.User = common.HexToAddress(log.Topics[1].Hex())
	event.Token = common.HexToAddress(log.Topics[2].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[3].Bytes())

	return event
}
