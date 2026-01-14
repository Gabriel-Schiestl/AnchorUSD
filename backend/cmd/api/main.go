package main

import (
	"log"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/blockchain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/config"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.GetDBInstance()
	bChainConfig := config.GetBlockchainConfig()
	bChainClient := blockchain.GetClient(bChainConfig)

	metricsStore := storage.NewMetricsStore(db)
	metricsService := service.NewMetricsService(metricsStore)

	worker.RunLogWorker(bChainClient)
	http.RegisterRoutes(metricsService)
	http.Run(":8080")
}