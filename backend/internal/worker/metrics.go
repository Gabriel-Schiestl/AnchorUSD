package worker

import (
	"os"
	"strconv"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service/processors"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type metricsProcessor struct {
	cacheStore storage.ICacheStore
}

var metricsChan chan model.Metrics

func init() {
	metricsChan = make(chan model.Metrics, 500)
}

func RunMetricsWorker(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting metrics worker")

	numMetricsWorkers := os.Getenv("NUM_METRICS_WORKERS")
	if numMetricsWorkers == "" {
		numMetricsWorkers = "4"
	}

	intNumMetricsWorkers, err := strconv.Atoi(numMetricsWorkers)
	if err != nil {
		logger.Warn().Err(err).Msg("Invalid NUM_METRICS_WORKERS value, defaulting to 4")
		intNumMetricsWorkers = 4
	}

	logger.Info().Int("workers", intNumMetricsWorkers).Msg("Starting metrics processing workers")
	for i := 0; i < intNumMetricsWorkers; i++ {
		mp := &metricsProcessor{cacheStore: cacheStore}
		logger.Debug().Int("worker_id", i+1).Msg("Starting metrics worker")
		go mp.process(cacheStore, priceFeed, priceStore)
	}
	logger.Info().Msg("All metrics workers started successfully")
}

func (mp *metricsProcessor) process(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	logger := utils.GetLogger()
	logger.Debug().Msg("Metrics processor goroutine started")

	for metric := range metricsChan {
		logger.Debug().Str("asset", string(metric.Asset)).Str("user", metric.UserAddress.Hex()).Msg("Processing metric from channel")

		switch metric.Asset {
		case model.CollateralAsset:
			logger.Debug().Str("user", metric.UserAddress.Hex()).Msg("Processing collateral metric")
			processors.ProcessCollateral(metric, cacheStore, priceFeed, priceStore)
			logger.Info().Str("user", metric.UserAddress.Hex()).Msg("Collateral metric processed successfully")

		case model.StablecoinAsset:
			logger.Debug().Str("user", metric.UserAddress.Hex()).Msg("Processing stablecoin metric")
			processors.ProcessCoin(metric, cacheStore)
			logger.Info().Str("user", metric.UserAddress.Hex()).Msg("Stablecoin metric processed successfully")

		default:
			logger.Warn().Str("asset", string(metric.Asset)).Str("user", metric.UserAddress.Hex()).Msg("Unknown asset type in metric")
		}
	}
}
