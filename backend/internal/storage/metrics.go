package storage

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
)

type metricsStore struct {
	DB *gorm.DB
}

func NewMetricsStore(db *gorm.DB) *metricsStore {
	return &metricsStore{
		DB: db,
	}
}

func (s *metricsStore) Get(ctx context.Context) (*model.Metrics, error) {
	// Implementation to retrieve metrics from the database
	return &model.Metrics{}, nil
}
