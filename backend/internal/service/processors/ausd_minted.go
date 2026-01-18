package processors

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func ProcessAUSDMinted(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	event := decodeAUSDMintedEvent(log)
	if event == nil {
		return
	}

	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.Hex(),
		LogIndex:    log.Index,
		Name:        eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		return
	}

	mint := &model.Mints{
		ID:          uuid.New().String(),
		EventID:     eventModel.ID,
		UserAddress: event.To.Hex(),
		Amount:      event.Amount,
	}

	err = storage.GetCoinStore().CreateMint(context.Background(), mint)
	if err != nil {
		return
	}

	metricsChan <- model.Metrics{
		UserAddress: event.To,
		Amount:      event.Amount,
		Asset:       model.StablecoinAsset,
		Operation:   model.Addition,
		BlockNumber: eventModel.BlockNumber,
	}
}

func decodeAUSDMintedEvent(log types.Log) *model.AUSDMintedEvent {
	// AUSDMinted(address,uint256) - address e uint256 são indexed
	if len(log.Topics) < 3 {
		return nil
	}

	event := &model.AUSDMintedEvent{}
	event.To = common.HexToAddress(log.Topics[1].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[2].Bytes())

	return event
}
