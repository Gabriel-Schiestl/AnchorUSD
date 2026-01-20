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
	store = eventsStore{DB: db}
	return &store
}

func GetEventsStore() *eventsStore {
	return &store
}

func (s *eventsStore) FindOneInBlock(ctx context.Context, logId uint, blockNumber uint64) (*model.Events, error) {
	var event model.Events
	result := s.DB.WithContext(ctx).Where("log_index = ? AND block_number = ?", logId, blockNumber).First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &event, nil
}

func (s *eventsStore) Create(ctx context.Context, event *model.Events) error {
	result := s.DB.WithContext(ctx).Create(event)
	return result.Error
}

func (s *eventsStore) GetLastProcessedBlock() (int64, error) {
	var event model.Events
	result := s.DB.Order("block_number desc").First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	return int64(event.BlockNumber), nil
}
