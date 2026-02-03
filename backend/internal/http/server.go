package http

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/middlewares"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

var server *gin.Engine

func init() {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing Gin HTTP server")
	server = gin.Default()
	server.Use(middlewares.CORSMiddleware())
	server.Use(middlewares.PrometheusMiddleware())

	limiter := rate.NewLimiter(2, 10)
	server.Use(middlewares.RateLimitMiddleware(limiter))

	server.GET("/metrics", gin.WrapH(promhttp.Handler()))
	logger.Info().Msg("Prometheus metrics endpoint registered at /metrics")
}

func Run(addr string) error {
	logger := utils.GetLogger()
	logger.Info().Str("address", addr).Msg("Starting HTTP server")
	err := server.Run(addr)
	if err != nil {
		logger.Fatal().Err(err).Str("address", addr).Msg("Failed to start HTTP server")
	}
	return err
}
