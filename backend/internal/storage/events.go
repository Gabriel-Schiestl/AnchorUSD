package storage

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"gorm.io/gorm"
)

var store eventsStore

type eventsStore struct {
	DB *gorm.DB
}

func NewEventsStore(db *gorm.DB) *eventsStore {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing events store")
	store = eventsStore{DB: db}
	return &store
}

func GetEventsStore() *eventsStore {
	return &store
}

func (s *eventsStore) FindOneInBlock(ctx context.Context, logId uint, blockNumber uint64) (*model.Events, error) {
	logger := utils.GetLogger()
	logger.Debug().Uint("log_id", logId).Uint64("block", blockNumber).Msg("Searching for event in block")

	var event model.Events
	result := s.DB.WithContext(ctx).Where("log_index = ? AND block_number = ?", logId, blockNumber).First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Debug().Uint("log_id", logId).Uint64("block", blockNumber).Msg("Event not found in block")
			return nil, nil
		}
		logger.Error().Err(result.Error).Uint("log_id", logId).Uint64("block", blockNumber).Msg("Error finding event in block")
		return nil, result.Error
	}
	logger.Debug().Uint("event_id", event.ID).Msg("Event found in database")
	return &event, nil
}

func (s *eventsStore) Create(ctx context.Context, event *model.Events) error {
	logger := utils.GetLogger()
	logger.Debug().Str("event_name", event.Name).Uint64("block", event.BlockNumber).Msg("Creating event in database")

	result := s.DB.WithContext(ctx).Create(event)
	if result.Error != nil {
		logger.Error().Err(result.Error).Str("event_name", event.Name).Msg("Failed to create event")
	}
	return result.Error
}

func (s *eventsStore) GetLastProcessedBlock() (int64, error) {
	logger := utils.GetLogger()
	logger.Debug().Msg("Getting last processed block")

	var event model.Events
	result := s.DB.Order("block_number desc").First(&event)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Debug().Msg("No events found, returning block 0")
			return 0, nil
		}
		logger.Error().Err(result.Error).Msg("Error getting last processed block")
		return 0, result.Error
	}
	logger.Info().Int64("last_block", int64(event.BlockNumber)).Msg("Last processed block retrieved")
	return int64(event.BlockNumber), nil
}
