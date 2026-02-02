package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ausd_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	TotalCollateralUSD = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ausd_total_collateral_usd",
			Help: "Total collateral locked in USD by token type",
		},
		[]string{"token"},
	)

	CollateralDepositsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_collateral_deposits_total",
			Help: "Total number of collateral deposits",
		},
		[]string{"token"},
	)

	CollateralRedeemsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_collateral_redeems_total",
			Help: "Total number of collateral redeems",
		},
		[]string{"token"},
	)

	AUSDTotalSupply = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_total_supply",
			Help: "Total AUSD supply in circulation",
		},
	)

	AUSDMintsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_mints_total",
			Help: "Total number of AUSD mints",
		},
	)

	AUSDBurnsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_burns_total",
			Help: "Total number of AUSD burns",
		},
	)

	AUSDMintedAmount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_minted_amount_total",
			Help: "Total amount of AUSD minted",
		},
	)

	AUSDBurnedAmount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_burned_amount_total",
			Help: "Total amount of AUSD burned",
		},
	)

	ActiveUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_active_users_total",
			Help: "Total number of active users with positions",
		},
	)

	AverageHealthFactor = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_average_health_factor",
			Help: "Average health factor across all users",
		},
	)

	UsersAtRisk = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_users_at_risk",
			Help: "Number of users with health factor below safe threshold",
		},
	)

	LiquidatableUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_liquidatable_users",
			Help: "Number of users currently liquidatable",
		},
	)

	LiquidationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_liquidations_total",
			Help: "Total number of liquidations executed",
		},
	)

	LiquidatedDebtTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_liquidated_debt_total",
			Help: "Total debt liquidated in AUSD",
		},
	)

	CollateralizationRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_collateralization_ratio",
			Help: "Protocol-wide collateralization ratio",
		},
	)

	BackingPercentage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ausd_backing_percentage",
			Help: "Percentage of AUSD backed by collateral",
		},
	)

	TokenPriceUSD = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ausd_token_price_usd",
			Help: "Current price of collateral tokens in USD",
		},
		[]string{"token"},
	)

	PriceFeedErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_price_feed_errors_total",
			Help: "Total number of price feed errors",
		},
		[]string{"token"},
	)

	OperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ausd_operation_duration_seconds",
			Help:    "Duration of blockchain operations in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)

	OperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_operation_errors_total",
			Help: "Total number of operation errors",
		},
		[]string{"operation", "error_type"},
	)

	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "ausd_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ausd_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ausd_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)
)

func RecordHTTPRequest(method, endpoint string, statusCode int) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(statusCode))).Inc()
}

func RecordOperation(operation string, duration float64) {
	OperationDuration.WithLabelValues(operation).Observe(duration)
}

func RecordError(operation, errorType string) {
	OperationErrors.WithLabelValues(operation, errorType).Inc()
}
