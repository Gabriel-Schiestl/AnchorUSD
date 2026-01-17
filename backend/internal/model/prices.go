package model

type Prices struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	TokenName   string `gorm:"size:4;not null"`
	TokenAddress string `gorm:"index:idx_token_block,unique;size:42;not null"`
	BlockNumber  uint64 `gorm:"index:idx_token_block,unique;not null"`
	PriceInUSD   string `gorm:"type:numeric(78,0);not null"`
}