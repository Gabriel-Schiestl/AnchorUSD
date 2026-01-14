package worker

import (
	"context"
	"log"
	"math/big"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service/processors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainConfig interface {
	GetContractAddress() string
}

type EventStore interface {
	GetLastProcessedBlock() (int64, error)
}

type Processor func(types.Log)

var Processors = map[string]Processor{
	"CollateralDeposited": processors.ProcessCollateralDeposited,
	"AUSDMinted":         processors.ProcessAUSDMinted,
	"AUSDBurned":         processors.ProcessAUSDBurned,
	"CollateralRedeemed": processors.ProcessCollateralRedeemed,
}

func RunLogWorker(bchainClient *ethclient.Client, bchainConfig BlockchainConfig, eventStore EventStore) {
	lastEventBlock, err := eventStore.GetLastProcessedBlock()
	if err != nil {
		log.Fatal(err)
	}

	if lastEventBlock == 0 {
		lastEventBlock = 1
	}

	logsChan := make(chan types.Log)

	sub, err := bchainClient.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: big.NewInt(lastEventBlock),
		ToBlock:   nil,
		Addresses: []common.Address{common.HexToAddress(bchainConfig.GetContractAddress())},
	}, logsChan)
	if err != nil {
		panic(err)
	}
	
	go processLogs(logsChan, sub)
}

func processLogs(logsChan <-chan types.Log, sub ethereum.Subscription) {
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logsChan:
			decodeLog(vLog)
		}
	}
}

func decodeLog(vLog types.Log) {
	for _, event := range model.EventsSignatures {
		if event.MatchesHexSignature(vLog.Topics[0].Hex()) {
			log.Printf("Matched topic %s", event.GetName())
			checkProcessorExists(event.GetName(), vLog)
		}
	} 
}

func checkProcessorExists(eventName string, vLog types.Log) {
	if processor, exists := Processors[eventName]; exists {
		go processor(vLog)
	} 

	log.Printf("No processor found for event: %s", eventName)
}