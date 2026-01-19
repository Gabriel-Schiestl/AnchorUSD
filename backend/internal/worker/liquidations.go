package worker

import (
	"os"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func RunLiquidationsWorker(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	tickerScanInterval := os.Getenv("LIQUIDATIONS_SCAN_INTERVAL")
	if tickerScanInterval == "" {
		tickerScanInterval = "1h"
	}

	duration, err := time.ParseDuration(tickerScanInterval)
	if err != nil {
		duration = time.Hour
	}

	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ticker.C:
				service.CalculateLiquidations(priceFeed, cacheStore)
			}
		}
	}()
}
