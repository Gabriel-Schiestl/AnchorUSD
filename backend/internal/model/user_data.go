package model

type UserData struct {
	TotalDebt string `json:"total_debt"`
	CollateralValueUSD string `json:"collateral_value_usd"`
	MaxMintable string `json:"max_mintable"`
	CurrentHealthFactor string `json:"current_health_factor"`
}