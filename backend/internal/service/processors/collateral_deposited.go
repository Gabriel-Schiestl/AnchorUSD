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

func ProcessCollateralDeposited(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	event := decodeCollateralDepositedEvent(log)
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

	deposit := &model.Deposit{
		ID:                uuid.New().String(),
		EventID:           eventModel.ID,
		UserAddress:       event.From.Hex(),
		CollateralAddress: event.TokenAddr.Hex(),
		Amount:            event.Amount,
	}

	err = storage.GetCollateralStore().CreateDeposit(context.Background(), deposit)
	if err != nil {
		return
	}

	metricsChan <- model.Metrics{
		UserAddress: event.From,
		Amount:      event.Amount,
		Asset:       model.CollateralAsset,
		Operation:   model.Addition,
		BlockNumber: eventModel.BlockNumber,
	}
}

func decodeCollateralDepositedEvent(log types.Log) *model.CollateralDepositedEvent {
	// CollateralDeposited(address,address,uint256) - todos são indexed
	if len(log.Topics) < 4 {
		return nil
	}

	event := &model.CollateralDepositedEvent{}
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.TokenAddr = common.HexToAddress(log.Topics[2].Hex())
	event.Amount = new(big.Int).SetBytes(log.Topics[3].Bytes())

	return event
}
