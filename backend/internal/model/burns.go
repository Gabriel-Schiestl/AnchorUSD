package model

import "github.com/shopspring/decimal"

type Burns struct {
	ID          string          `json:"id" gorm:"primaryKey"`
	EventID     uint            `json:"event_id" gorm:"event_id"`
	UserAddress string          `json:"user_address" gorm:"user_address"`
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(78,0);amount"`
}
