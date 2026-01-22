package http

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/handlers"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	svc handlers.MetricsReader,
	userDataSvc handlers.UserReader,
	hfCalcSvc handlers.HealthFactorCalculator,
	dashboardMetricsSvc handlers.DashboardMetricsReader,
) {
	logger := utils.GetLogger()
	logger.Info().Msg("Registering HTTP routes")

	api := server.Group("/api")

	api.GET("/status", func(c *gin.Context) {
		logger.Debug().Msg("Health check endpoint called")
		c.JSON(200, gin.H{"status": "ok"})
	})
	logger.Debug().Msg("Registered /api/status route")
	
	metrics := api.Group("/metrics")
	{
		metrics.GET("", handlers.GetMetricsHandler(svc))
		metrics.GET("/dashboard", handlers.GetDashboardMetricsHandler(dashboardMetricsSvc))
	}
	logger.Debug().Msg("Registered /api/metrics routes")

	api.GET("/user/:user", handlers.GetUserDataHandler(userDataSvc))
	logger.Debug().Msg("Registered /api/user/:user route")

	ausdEngine := api.Group("/ausd-engine")
	{
		ausdEngine.POST("/calculate-mint", handlers.CalculateMintHandler(hfCalcSvc))
		ausdEngine.POST("/calculate-burn", handlers.CalculateBurnHandler(hfCalcSvc))
		ausdEngine.POST("/calculate-deposit", handlers.CalculateDepositHandler(hfCalcSvc))
	}
	logger.Debug().Msg("Registered /api/ausd-engine routes")

	logger.Info().Msg("All HTTP routes registered successfully")
}
