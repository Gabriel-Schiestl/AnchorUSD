package model

type Deposit struct {
	ID          string  `json:"id" gorm:"primaryKey"`
	EventID     uint    `json:"event_id" gorm:"event_id"`
	UserAddress string  `json:"user_address" gorm:"user_address"`
	CollateralAddress string  `json:"collateral_address" gorm:"collateral_address"`
	Amount      float64 `json:"amount" gorm:"amount"`
}