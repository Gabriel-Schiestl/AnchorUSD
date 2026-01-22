package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type DashboardMetricsReader interface {
	GetDashboardMetrics(ctx context.Context) (model.DashboardMetrics, error)
}

func GetDashboardMetricsHandler(svc DashboardMetricsReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		logger.Info().Str("endpoint", "/dashboard/metrics").Msg("Request received for dashboard metrics")

		metrics, err := svc.GetDashboardMetrics(ctx.Request.Context())
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get dashboard metrics")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Int("liquidatable_users", len(metrics.LiquidatableUsers)).Str("total_collateral", metrics.TotalCollateral.Value).Msg("Dashboard metrics retrieved successfully")
		ctx.JSON(200, metrics)
	}
}
