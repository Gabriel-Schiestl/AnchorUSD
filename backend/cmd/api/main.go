package main

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
)

func main() {
	metricsStore := storage.NewMetricsStore(nil)
	metricsService := service.NewMetricsService(metricsStore)
	http.RegisterRoutes(metricsService)
	http.Run(":8080")
}