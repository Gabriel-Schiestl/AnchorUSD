package model

type Events struct {
	ID 	   uint    `json:"id" gorm:"primaryKey"`
	BlockNumber uint64  `json:"block_number" gorm:"block_number"`
	TxHash     string  `json:"tx_hash" gorm:"tx_hash"`
	LogIndex   uint    `json:"log_index" gorm:"log_index"`
	Name       string  `json:"name" gorm:"name"`
	CreatedAt  int64   `json:"created_at" gorm:"created_at"`
}