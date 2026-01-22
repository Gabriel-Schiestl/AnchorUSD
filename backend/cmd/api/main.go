package main

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/blockchain"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/config"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/http/external"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/service"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/storage"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/worker"

	"github.com/joho/godotenv"
)

func main() {
	logger := utils.GetLogger()
	logger.Info().Msg("Starting AnchorUSD Backend Application")

	logger.Debug().Msg("Loading environment variables from .env file")
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error loading .env file")
	}
	logger.Info().Msg("Environment variables loaded successfully")

	logger.Info().Msg("Loading collateral tokens configuration")
	constants.LoadCollateralTokens()
	logger.Info().Msg("Collateral tokens loaded successfully")

	logger.Info().Msg("Initializing database connection")
	db := config.GetDBInstance()
	logger.Info().Msg("Database connection established")

	logger.Info().Msg("Loading blockchain configuration")
	bChainConfig := config.GetBlockchainConfig()
	logger.Info().Str("contract_address", bChainConfig.GetContractAddress()).Msg("Blockchain configuration loaded")

	logger.Info().Msg("Loading cache configuration")
	cacheConfig := config.GetCacheConfig()
	logger.Info().Msg("Cache configuration loaded")

	logger.Info().Msg("Running database migrations")
	db.AutoMigrate(model.Events{}, model.Burns{}, model.Deposit{}, model.Events{}, model.Mints{}, model.Prices{}, model.Redeem{})
	logger.Info().Msg("Database migrations completed successfully")

	logger.Info().Msg("Initializing blockchain client")
	bChainClient := blockchain.GetClient(bChainConfig)
	logger.Info().Msg("Blockchain client initialized")

	logger.Info().Msg("Initializing cache store")
	cacheStore := storage.NewCacheStore(cacheConfig)
	logger.Info().Msg("Cache store initialized")

	logger.Info().Msg("Initializing metrics store and service")
	metricsStore := storage.NewMetricsStore(db)
	metricsService := service.NewMetricsService(metricsStore)
	logger.Info().Msg("Metrics service ready")

	logger.Info().Msg("Initializing user data service")
	userDataService := service.NewUserDataService(cacheStore)
	logger.Info().Msg("User data service ready")

	logger.Info().Msg("Initializing price feed API")
	priceFeed := external.NewPriceFeedAPI()
	logger.Info().Msg("Price feed API initialized")

	logger.Info().Msg("Initializing health factor calculation service")
	healthFactorCalcService := service.NewHealthFactorCalculationService(cacheStore, priceFeed)
	logger.Info().Msg("Health factor calculation service ready")

	logger.Info().Msg("Initializing dashboard metrics service")
	dashboardMetricsService := service.NewDashboardMetricsService(cacheStore, priceFeed)
	logger.Info().Msg("Dashboard metrics service ready")

	logger.Info().Msg("Initializing storage layers")
	eventStore := storage.NewEventsStore(db)
	storage.NewCoinStore(db)
	storage.NewCollateralStore(db)
	priceStore := storage.NewPriceStore(db)
	logger.Info().Msg("All storage layers initialized")

	logger.Info().Msg("Starting log worker for blockchain events")
	worker.RunLogWorker(bChainClient, bChainConfig, eventStore)
	logger.Info().Msg("Log worker started")

	logger.Info().Msg("Starting metrics worker")
	worker.RunMetricsWorker(cacheStore, priceFeed, priceStore)
	logger.Info().Msg("Metrics worker started")

	logger.Info().Msg("Starting liquidations worker")
	worker.RunLiquidationsWorker(cacheStore, priceFeed)
	logger.Info().Msg("Liquidations worker started")

	logger.Info().Msg("Updating initial metrics")
	service.UpdateMetrics(cacheStore, priceFeed)
	logger.Info().Msg("Initial metrics updated")

	logger.Info().Msg("Registering HTTP routes")
	http.RegisterRoutes(metricsService, userDataService, healthFactorCalcService, dashboardMetricsService)
	logger.Info().Msg("HTTP routes registered")

	logger.Info().Str("address", ":8080").Msg("Starting HTTP server")
	http.Run(":8080")
}
