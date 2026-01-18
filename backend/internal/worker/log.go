package worker

import (
	"context"
	"log"
	"math/big"
	"os"
	"strconv"

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
	FindOneInBlock(ctx context.Context, logId uint, blockNumber uint64) (*model.Events, error)
}

type Processor func(string, types.Log, chan<- model.Metrics)

var Processors = map[string]Processor{
	"CollateralDeposited": processors.ProcessCollateralDeposited,
	"AUSDMinted":          processors.ProcessAUSDMinted,
	"AUSDBurned":          processors.ProcessAUSDBurned,
	"CollateralRedeemed":  processors.ProcessCollateralRedeemed,
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

	numLogWorkers := os.Getenv("NUM_LOG_WORKERS")
	if numLogWorkers == "" {
		numLogWorkers = "4"
	}

	intNumLogWorkers, err := strconv.Atoi(numLogWorkers)
	if err != nil {
		intNumLogWorkers = 4
	}

	for i := 0; i < intNumLogWorkers; i++ {
		go processLogs(logsChan, sub, eventStore)
	}
}

func processLogs(logsChan <-chan types.Log, sub ethereum.Subscription, eventStore EventStore) {
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logsChan:
			decodeLog(vLog, eventStore)
		}
	}
}

func decodeLog(vLog types.Log, eventStore EventStore) {
	if event, _ := eventStore.FindOneInBlock(context.Background(), vLog.Index, vLog.BlockNumber); event != nil {
		return
	}

	for _, event := range model.EventsSignatures {
		if event.MatchesHexSignature(vLog.Topics[0].Hex()) {
			log.Printf("Matched topic %s", event.GetName())
			checkProcessorExists(event.GetName(), vLog)
		}
	}
}

func checkProcessorExists(eventName string, vLog types.Log) {
	if processor, exists := Processors[eventName]; exists {
		go processor(eventName, vLog, metricsChan)
	}

	log.Printf("No processor found for event: %s", eventName)
}
