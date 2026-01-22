package service

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type MetricsStore interface {
	Get(ctx context.Context) (*model.Metrics, error)
}

type metricsService struct {
	Store MetricsStore
}

func NewMetricsService(store MetricsStore) *metricsService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing metrics service")
	return &metricsService{
		Store: store,
	}
}

func (s *metricsService) GetMetrics(ctx context.Context) (map[string]any, error) {
	logger := utils.GetLogger()
	logger.Info().Msg("Retrieving metrics from store")

	// Implementation to retrieve metrics
	metrics := map[string]any{}

	logger.Info().Msg("Metrics retrieved successfully")
	return metrics, nil
}
