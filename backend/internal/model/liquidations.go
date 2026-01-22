package model

type Liquidations struct {
	ID                    string `json:"id" gorm:"primaryKey"`
	EventID               uint   `json:"event_id" gorm:"event_id"`
	LiquidatedUserAddress string `json:"liquidated_user_address" gorm:"liquidated_user_address"`
	LiquidatorAddress     string `json:"liquidator_address" gorm:"liquidator_address"`
	CollateralAddress     string `json:"collateral_address" gorm:"collateral_address"`
	CollateralAmount      BigInt `json:"collateral_amount" gorm:"type:numeric"`
	DebtCovered           BigInt `json:"debt_covered" gorm:"type:numeric"`
}
