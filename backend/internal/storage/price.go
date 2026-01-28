package storage

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPriceStore interface {
	GetPriceInBlock(tokenAddress string, blockNumber uint64) (*string, error)
	SavePriceInBlock(tokenAddress string, blockNumber uint64, priceInUSD string) error
}

var priceStr priceStore

type priceStore struct {
	DB *gorm.DB
}

func NewPriceStore(db *gorm.DB) *priceStore {
	return &priceStore{
		DB: db,
	}
}

func GetPriceStore() *priceStore {
	return &priceStr
}

func (s *priceStore) GetPriceInBlock(tokenName string, blockNumber uint64) (*string, error) {
	var price model.Prices
	result := s.DB.Where("token_name = ? AND block_number <= ?", tokenName, blockNumber).Order("block_number desc").First(&price)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &price.PriceInUSD, nil
}

func (s *priceStore) SavePriceInBlock(tokenName string, blockNumber uint64, priceInUSD string) error {
	price := model.Prices{
		TokenName:   tokenName,
		BlockNumber: blockNumber,
		PriceInUSD:  priceInUSD,
	}
	result := s.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&price)
	return result.Error
}
