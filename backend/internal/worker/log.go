package worker

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service/processors"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
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
	"Liquidation":         processors.ProcessLiquidation,
}

func RunLogWorker(bchainClient *ethclient.Client, bchainConfig BlockchainConfig, eventStore EventStore) {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting blockchain log worker")

	lastEventBlock, err := eventStore.GetLastProcessedBlock()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get last processed block")
	}

	if lastEventBlock == 0 {
		lastEventBlock = 1
		logger.Debug().Msg("No previous blocks processed, starting from block 1")
	}
	logger.Info().Int64("from_block", lastEventBlock).Msg("Starting from last processed block")

	logsChan := make(chan types.Log)

	contractAddr := bchainConfig.GetContractAddress()
	logger.Info().Str("contract_address", contractAddr).Msg("Subscribing to contract logs")

	sub, err := bchainClient.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: big.NewInt(lastEventBlock),
		ToBlock:   nil,
		Addresses: []common.Address{common.HexToAddress(contractAddr)},
	}, logsChan)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to subscribe to filter logs")
	}
	logger.Info().Msg("Successfully subscribed to blockchain logs")

	numLogWorkers := os.Getenv("NUM_LOG_WORKERS")
	if numLogWorkers == "" {
		numLogWorkers = "4"
	}

	intNumLogWorkers, err := strconv.Atoi(numLogWorkers)
	if err != nil {
		logger.Warn().Err(err).Msg("Invalid NUM_LOG_WORKERS value, defaulting to 4")
		intNumLogWorkers = 4
	}

	logger.Info().Int("workers", intNumLogWorkers).Msg("Starting log processing workers")
	for i := 0; i < intNumLogWorkers; i++ {
		logger.Debug().Int("worker_id", i+1).Msg("Starting log worker")
		go processLogs(logsChan, sub, eventStore)
	}
	logger.Info().Msg("All log workers started successfully")
}

func processLogs(logsChan <-chan types.Log, sub ethereum.Subscription, eventStore EventStore) {
	logger := utils.GetLogger()
	logger.Debug().Msg("Log processing goroutine started")

	for {
		select {
		case err := <-sub.Err():
			logger.Fatal().Err(err).Msg("Blockchain subscription error")
		case vLog := <-logsChan:
			logger.Debug().Uint64("block", vLog.BlockNumber).Uint("index", vLog.Index).Msg("Received log from blockchain")
			decodeLog(vLog, eventStore)
		}
	}
}

func decodeLog(vLog types.Log, eventStore EventStore) {
	logger := utils.GetLogger()
	logger.Info().Uint64("block", vLog.BlockNumber).Uint("index", vLog.Index).Msg("Processing blockchain log")

	fmt.Println("Processing log in block:", vLog.BlockNumber, " with index:", vLog.Index)

	if event, _ := eventStore.FindOneInBlock(context.Background(), vLog.Index, vLog.BlockNumber); event != nil {
		logger.Debug().Uint64("block", vLog.BlockNumber).Uint("index", vLog.Index).Msg("Event already processed, skipping")
		return
	}

	for _, event := range model.EventsSignatures {
		if event.MatchesHexSignature(vLog.Topics[0].Hex()) {
			eventName := event.GetName()
			logger.Info().Str("event", eventName).Uint64("block", vLog.BlockNumber).Uint("index", vLog.Index).Msg("Event signature matched")
			checkProcessorExists(eventName, vLog)
		}
	}
}

func checkProcessorExists(eventName string, vLog types.Log) {
	logger := utils.GetLogger()
	if processor, exists := Processors[eventName]; exists {
		logger.Debug().Str("event", eventName).Msg("Processor found, processing event")
		processor(eventName, vLog, metricsChan)
		logger.Info().Str("event", eventName).Uint64("block", vLog.BlockNumber).Msg("Event processed and sent to metrics channel")
	} else {
		logger.Warn().Str("event", eventName).Msg("No processor found for event")
	}
}
