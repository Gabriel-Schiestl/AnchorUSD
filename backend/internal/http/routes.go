package http

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(svc handlers.MetricsReader) {
	api := server.Group("/api")

	api.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api.GET("/metrics", handlers.GetMetricsHandler(svc))
}