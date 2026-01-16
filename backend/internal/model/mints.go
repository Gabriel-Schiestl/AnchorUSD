package model

import (
	"math/big"
)

type Mints struct {
	ID          string          `json:"id" gorm:"primaryKey"`
	EventID     uint            `json:"event_id" gorm:"event_id"`
	UserAddress string          `json:"user_address" gorm:"user_address"`
	Amount      *big.Int        `json:"amount" gorm:"amount"`
}
