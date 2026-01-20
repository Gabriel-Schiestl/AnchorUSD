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

func RunMetricsWorker(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	numMetricsWorkers := os.Getenv("NUM_METRICS_WORKERS")
	if numMetricsWorkers == "" {
		numMetricsWorkers = "4"
	}

	intNumMetricsWorkers, err := strconv.Atoi(numMetricsWorkers)
	if err != nil {
		intNumMetricsWorkers = 4
	}

	for i := 0; i < intNumMetricsWorkers; i++ {
		mp := &metricsProcessor{cacheStore: cacheStore}
		go mp.process(cacheStore, priceFeed, priceStore)
	}
}

func (mp *metricsProcessor) process(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI, priceStore storage.IPriceStore) {
	for metric := range metricsChan {
		switch metric.Asset {
		case model.CollateralAsset:
			processors.ProcessCollateral(metric, cacheStore, priceFeed, priceStore)
		case model.StablecoinAsset:
			processors.ProcessCoin(metric, cacheStore)
		}
	}
}
