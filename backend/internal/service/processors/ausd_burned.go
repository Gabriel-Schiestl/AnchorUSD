package processors

import (
	"context"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func ProcessAUSDBurned(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	event := decodeAUSDBurnedEvent(log)
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

	amountDecimal := decimal.NewFromBigInt(event.Amount, 0)

	burn := &model.Burns{
		ID:          uuid.New().String(),
		EventID:     eventModel.ID,
		UserAddress: event.From.Hex(),
		Amount:      amountDecimal,
	}

	err = storage.GetCollateralStore().CreateBurn(context.Background(), burn)
	if err != nil {
		return
	}

	metricsChan <- model.Metrics{
		UserAddress: event.From,
		Amount: amountDecimal,
		Asset: model.StablecoinAsset,
		Operation: model.Subtraction,
	}
}

func decodeAUSDBurnedEvent(log types.Log) *model.AUSDBurnedEvent {
	if len(log.Topics) < 3 {
		return nil
	}

	event := &model.AUSDBurnedEvent{}
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[2].Bytes())

	return event
}