package processors

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/metrics"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func ProcessCollateralDeposited(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	logger := utils.GetLogger()
	logger.Info().Str("event", eventName).Uint64("block", log.BlockNumber).Uint("index", log.Index).Msg("Processing collateral deposited event")

	event := decodeCollateralDepositedEvent(log)
	if event == nil {
		logger.Error().Str("event", eventName).Uint64("block", log.BlockNumber).Msg("Failed to decode collateral deposited event")
		metrics.RecordError("deposit", "decode_error")
		return
	}

	logger.Debug().Str("user", event.From.Hex()).Str("token", event.TokenAddr.Hex()).Str("amount", event.Amount.String()).Msg("Event decoded successfully")

	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		logger.Error().Err(err).Str("event", eventName).Msg("Failed to create event in database")
		metrics.RecordError("deposit", "database_error")
		return
	}
	logger.Debug().Uint("event_id", eventModel.ID).Msg("Event record created in database")

	deposit := &model.Deposit{
		ID:                uuid.New().String(),
		EventID:           eventModel.ID,
		UserAddress:       event.From.Hex(),
		CollateralAddress: event.TokenAddr.Hex(),
		Amount:            model.NewBigInt(event.Amount),
	}

	err = storage.GetCollateralStore().CreateDeposit(context.Background(), deposit)
	if err != nil {
		logger.Error().Err(err).Str("user", event.From.Hex()).Msg("Failed to create deposit record")
		metrics.RecordError("deposit", "database_error")
		return
	}
	logger.Debug().Str("deposit_id", deposit.ID).Msg("Deposit record created")

	tokenName := getTokenNameByAddress(event.TokenAddr.Hex())
	if tokenName != "" {
		metrics.CollateralDepositsTotal.WithLabelValues(tokenName).Inc()
	}

	metric := model.Metrics{
		UserAddress: event.From,
		Amount:      event.Amount,
		Asset:       model.CollateralAsset,
		Operation:   model.Addition,
		BlockNumber: eventModel.BlockNumber,
		CollateralTokenAddress: event.TokenAddr,
	}

	metricsChan <- metric
	logger.Info().Str("user", event.From.Hex()).Str("token", event.TokenAddr.Hex()).Str("amount", event.Amount.String()).Msg("Collateral deposited event processed and metric sent to channel")
}

func getTokenNameByAddress(address string) string {
	for name, addr := range constants.CollateralTokens {
		if addr == address {
			return name
		}
	}
	return ""
}

func decodeCollateralDepositedEvent(log types.Log) *model.CollateralDepositedEvent {
	if len(log.Topics) < 4 {
		return nil
	}

	event := &model.CollateralDepositedEvent{}
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.TokenAddr = common.HexToAddress(log.Topics[2].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[3].Bytes())
	fmt.Println("Decoded CollateralDeposited event:", event.Amount, event.From.Hex(), event.TokenAddr.Hex())
	return event
}
