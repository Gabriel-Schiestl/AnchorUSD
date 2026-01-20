package http

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(svc handlers.MetricsReader, userDataSvc handlers.UserReader, hfCalcSvc handlers.HealthFactorCalculator) {
	api := server.Group("/api")

	api.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api.GET("/metrics", handlers.GetMetricsHandler(svc))
	api.GET("/user/:user", handlers.GetUserDataHandler(userDataSvc))

	// Health factor calculation endpoints
	ausdEngine := api.Group("/ausd-engine")
	{
		ausdEngine.POST("/calculate-mint", handlers.CalculateMintHandler(hfCalcSvc))
		ausdEngine.POST("/calculate-burn", handlers.CalculateBurnHandler(hfCalcSvc))
		ausdEngine.POST("/calculate-deposit", handlers.CalculateDepositHandler(hfCalcSvc))
	}
}
