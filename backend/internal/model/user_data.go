package model

type CollateralDeposited struct {
	Asset    string `json:"asset"`
	Amount   string `json:"amount"`
	ValueUsd string `json:"valueUsd"`
}

type UserData struct {
	TotalDebt           string                `json:"total_debt"`
	CollateralValueUSD  string                `json:"collateral_value_usd"`
	MaxMintable         string                `json:"max_mintable"`
	CurrentHealthFactor string                `json:"current_health_factor"`
	CollateralDeposited []CollateralDeposited `json:"collateral_deposited"`
}