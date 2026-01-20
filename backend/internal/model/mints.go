package model

type Mints struct {
	ID          string   `json:"id" gorm:"primaryKey"`
	EventID     uint     `json:"event_id" gorm:"event_id"`
	UserAddress string   `json:"user_address" gorm:"user_address"`
	Amount      BigInt `json:"amount" gorm:"amount"`
}
