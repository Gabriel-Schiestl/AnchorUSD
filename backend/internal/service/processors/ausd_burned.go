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

func ProcessAUSDBurned(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	logger := utils.GetLogger()
	logger.Info().Str("event", eventName).Uint64("block", log.BlockNumber).Uint("index", log.Index).Msg("Processing AUSD burned event")

	event := decodeAUSDBurnedEvent(log)
	if event == nil {
		logger.Error().Str("event", eventName).Uint64("block", log.BlockNumber).Msg("Failed to decode AUSD burned event")
		metrics.RecordError("burn", "decode_error")
		return
	}

	logger.Debug().Str("user", event.User.Hex()).Str("amount", event.Amount.String()).Msg("AUSD burned event decoded successfully")

	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		logger.Error().Err(err).Str("event", eventName).Msg("Failed to create event in database")
		metrics.RecordError("burn", "database_error")
		return
	}
	logger.Debug().Uint("event_id", eventModel.ID).Msg("Event record created in database")

	burn := &model.Burns{
		ID:          uuid.New().String(),
		EventID:     eventModel.ID,
		UserAddress: event.User.Hex(),
		Amount:      model.NewBigInt(event.Amount),
	}

	err = storage.GetCoinStore().CreateBurn(context.Background(), burn)
	if err != nil {
		logger.Error().Err(err).Str("user", event.User.Hex()).Msg("Failed to create burn record")
		metrics.RecordError("burn", "database_error")
		return
	}
	logger.Debug().Str("burn_id", burn.ID).Msg("Burn record created")

	metrics.AUSDBurnsTotal.Inc()
	amountFloat := new(big.Float).SetInt(event.Amount)
	amountFloat64, _ := amountFloat.Float64()
	metrics.AUSDBurnedAmount.Add(amountFloat64 / 1e18)

	metric := model.Metrics{
		UserAddress: event.User,
		Amount:      event.Amount,
		Asset:       model.StablecoinAsset,
		Operation:   model.Subtraction,
		BlockNumber: eventModel.BlockNumber,
	}

	metricsChan <- metric
	logger.Info().Str("user", event.User.Hex()).Str("amount", event.Amount.String()).Msg("AUSD burned event processed and metric sent to channel")
}

func decodeAUSDBurnedEvent(log types.Log) *model.AUSDBurnedEvent {
	if len(log.Topics) < 3 {
		return nil
	}

	event := &model.AUSDBurnedEvent{}
	event.User = common.HexToAddress(log.Topics[1].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[2].Bytes())

	return event
}
