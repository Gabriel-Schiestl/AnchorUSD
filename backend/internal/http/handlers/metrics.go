package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type MetricsReader interface {
	GetMetrics(ctx context.Context) (map[string]any, error)
}

func GetMetricsHandler(svc MetricsReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		logger.Info().Str("endpoint", "/metrics").Msg("Request received for metrics")

		metrics, err := svc.GetMetrics(ctx.Request.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get metrics")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Interface("metrics", metrics).Msg("Metrics retrieved successfully")
		ctx.JSON(200, metrics)
	}
}
