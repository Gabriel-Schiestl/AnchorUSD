package worker

import (
	"os"
	"strconv"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/go-redis/redis"
)

type CacheStore interface {
	Get(key string) *redis.StringCmd
	Set(key string, value any, expiration time.Duration) *redis.StatusCmd
}

type metricsProcessor struct {
	cacheStore CacheStore
}

var metricsChan chan model.Metrics

func init() {
	metricsChan = make(chan model.Metrics, 500)
}

func RunMetricsWorker(cacheStore CacheStore) {
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
		go mp.process()
	}
}

func (mp *metricsProcessor) process() {
	for metric := range metricsChan {
		_ = metric // TODO: implement metrics processing
	}
}