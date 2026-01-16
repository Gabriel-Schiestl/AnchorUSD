package worker

import (
	"os"
	"strconv"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service/processors"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

type metricsProcessor struct {
	cacheStore storage.ICacheStore
}

var metricsChan chan model.Metrics

func init() {
	metricsChan = make(chan model.Metrics, 500)
}

func RunMetricsWorker(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	numLogWorkers := os.Getenv("NUM_LOG_WORKERS")
	if numLogWorkers == "" {
		numLogWorkers = "4"
	}

	intNumLogWorkers, err := strconv.Atoi(numLogWorkers)
	if err != nil {
		intNumLogWorkers = 4
	}

	for i := 0; i < intNumLogWorkers; i++ {
		mp := &metricsProcessor{cacheStore: cacheStore}
		go mp.process(cacheStore, priceFeed)
	}
}

func (mp *metricsProcessor) process(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	for metric := range metricsChan {
		switch metric.Asset {
		case model.CollateralAsset:
			processors.ProcessCollateral(metric, cacheStore, priceFeed)
		case model.StablecoinAsset:
			processors.ProcessCoin(metric, cacheStore)
		}
	}
}