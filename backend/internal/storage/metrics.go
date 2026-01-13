package storage

import (
	"context"
	"database/sql"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
)

type metricsStore struct {
	DB *sql.DB
}

func NewMetricsStore(db *sql.DB) *metricsStore {
	return &metricsStore{
		DB: db,
	}
}

func (s *metricsStore) Get(ctx context.Context) (*model.Metrics, error) {
	// Implementation to retrieve metrics from the database
	return &model.Metrics{}, nil
}