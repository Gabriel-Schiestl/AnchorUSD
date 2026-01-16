package model

import "math/big"

type Liquidations struct {
	ID                string  `json:"id" gorm:"primaryKey"`
	EventID           uint    `json:"event_id" gorm:"event_id"`
	LiquidatorAddress string  `json:"liquidator_address" gorm:"liquidator_address"`
	LiquidatedAddress string  `json:"liquidated_address" gorm:"liquidated_address"`
	Amount            *big.Int `json:"amount" gorm:"amount"`
}
