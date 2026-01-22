package processors

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func ProcessAUSDMinted(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	logger := utils.GetLogger()
	logger.Info().Str("event", eventName).Uint64("block", log.BlockNumber).Uint("index", log.Index).Msg("Processing AUSD minted event")

	event := decodeAUSDMintedEvent(log)
	if event == nil {
		logger.Error().Str("event", eventName).Uint64("block", log.BlockNumber).Msg("Failed to decode AUSD minted event")
		return
	}

	logger.Debug().Str("user", event.To.Hex()).Str("amount", event.Amount.String()).Msg("AUSD minted event decoded successfully")

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

	mint := &model.Mints{
		ID:          uuid.New().String(),
		EventID:     eventModel.ID,
		UserAddress: event.To.Hex(),
		Amount:      model.NewBigInt(event.Amount),
	}

	err = storage.GetCoinStore().CreateMint(context.Background(), mint)
	if err != nil {
		logger.Error().Err(err).Str("user", event.To.Hex()).Msg("Failed to create mint record")
		return
	}
	logger.Debug().Str("mint_id", mint.ID).Msg("Mint record created")

	metric := model.Metrics{
		UserAddress: event.To,
		Amount:      event.Amount,
		Asset:       model.StablecoinAsset,
		Operation:   model.Addition,
		BlockNumber: eventModel.BlockNumber,
	}

	metricsChan <- metric
	logger.Info().Str("user", event.To.Hex()).Str("amount", event.Amount.String()).Msg("AUSD minted event processed and metric sent to channel")
}

func decodeAUSDMintedEvent(log types.Log) *model.AUSDMintedEvent {
	if len(log.Topics) < 3 {
		return nil
	}

	event := &model.AUSDMintedEvent{}
	event.To = common.HexToAddress(log.Topics[1].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[2].Bytes())

	return event
}
