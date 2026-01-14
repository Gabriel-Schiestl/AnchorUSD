package service

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
)

type MetricsStore interface {
	Get(ctx context.Context) (*model.Metrics, error)
}

type metricsService struct {
	Store MetricsStore
}

func NewMetricsService(store MetricsStore) *metricsService {
	return &metricsService{
		Store: store,
	}
}

func (s *metricsService) GetMetrics(ctx context.Context) (map[string]any, error) {
	// Implementation to retrieve metrics
	return map[string]any{}, nil
}
