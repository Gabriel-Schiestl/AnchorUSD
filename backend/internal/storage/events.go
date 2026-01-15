package storage

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
)

var store eventsStore

type eventsStore struct {
	DB *gorm.DB
}

func NewEventsStore(db *gorm.DB) *eventsStore {
	return &eventsStore{
		DB: db,
	}
}

func GetEventsStore() *eventsStore {
	return &store
}

func (s *eventsStore) Get(ctx context.Context) (*model.Events, error) {
	// Implementation to retrieve events from the database
	return &model.Events{}, nil
}

func (s *eventsStore) Create(ctx context.Context, event *model.Events) error {
	result := s.DB.WithContext(ctx).Create(event)
	return result.Error
}
