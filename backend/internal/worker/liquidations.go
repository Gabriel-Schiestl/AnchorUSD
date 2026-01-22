package worker

import (
	"os"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

func RunLiquidationsWorker(cacheStore storage.ICacheStore, priceFeed external.IPriceFeedAPI) {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting liquidations worker")

	tickerScanInterval := os.Getenv("LIQUIDATIONS_SCAN_INTERVAL")
	if tickerScanInterval == "" {
		tickerScanInterval = "1h"
	}

	duration, err := time.ParseDuration(tickerScanInterval)
	if err != nil {
		logger.Warn().Err(err).Str("interval", tickerScanInterval).Msg("Invalid LIQUIDATIONS_SCAN_INTERVAL, defaulting to 1h")
		duration = time.Hour
	}

	logger.Info().Str("interval", duration.String()).Msg("Liquidations worker configured")

	ticker := time.NewTicker(duration)
	go func() {
		logger.Debug().Msg("Liquidations worker goroutine started")
		for {
			select {
			case <-ticker.C:
				logger.Info().Msg("Liquidations scan triggered")
				err := service.CalculateLiquidations(priceFeed, cacheStore)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to calculate liquidations")
				} else {
					logger.Info().Msg("Liquidations calculation completed successfully")
				}
			}
		}
	}()
	logger.Info().Msg("Liquidations worker started successfully")
}
