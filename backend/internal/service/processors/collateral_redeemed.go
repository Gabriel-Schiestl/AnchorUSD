package processors

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/blockchain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func ProcessCollateralRedeemed(eventName string, log types.Log, metricsChan chan<- model.Metrics) {
	event := decodeEventData(log)
	if event == nil {
		return
	}

	eventModel := &model.Events{
		BlockNumber: log.BlockNumber,
		TxHash: log.TxHash.Hex(),
		LogIndex: log.Index,
		Name: eventName,
	}

	err := storage.GetEventsStore().Create(context.Background(), eventModel)
	if err != nil {
		return
	}

	amountDecimal := decimal.NewFromBigInt(event.Amount, 0)

	collateral := &model.Redeem{
		ID:                uuid.New().String(),
		EventID:           eventModel.ID,
		UserAddress:       event.From.Hex(),
		CollateralAddress: event.TokenAddr.Hex(),
		Amount:            amountDecimal,
	}

	err = storage.GetCollateralStore().CreateRedeem(context.Background(), collateral)
	if err != nil {
		return
	}

	metricsChan <- model.Metrics{
		UserAddress: event.From,
		Amount: amountDecimal,
		Asset: model.CollateralAsset,
		Operation: model.Subtraction,
	}
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
	event.TokenAddr = common.HexToAddress(log.Topics[3].Hex())

	return event
}