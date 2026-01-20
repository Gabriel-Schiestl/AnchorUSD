package main

import (
	"log"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/blockchain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/config"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	constants.LoadCollateralTokens()

	db := config.GetDBInstance()
	bChainConfig := config.GetBlockchainConfig()
	cacheConfig := config.GetCacheConfig()

	db.AutoMigrate(model.Events{}, model.Burns{}, model.Deposit{}, model.Events{}, model.Mints{}, model.Prices{}, model.Redeem{})

	bChainClient := blockchain.GetClient(bChainConfig)
	cacheStore := storage.NewCacheStore(cacheConfig)

	metricsStore := storage.NewMetricsStore(db)
	metricsService := service.NewMetricsService(metricsStore)

	eventStore := storage.NewEventsStore(db)
	storage.NewCoinStore(db)
	storage.NewCollateralStore(db)
	

	priceStore := storage.NewPriceStore(db)

	priceFeed := external.NewPriceFeedAPI()

	worker.RunLogWorker(bChainClient, bChainConfig, eventStore)
	worker.RunMetricsWorker(cacheStore, priceFeed, priceStore)
	worker.RunLiquidationsWorker(cacheStore, priceFeed)

	service.UpdateMetrics(cacheStore, priceFeed)

	http.RegisterRoutes(metricsService)
	http.Run(":8080")
}
