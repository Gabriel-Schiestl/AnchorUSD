package model

type LiquidatableUser struct {
	Address           string `json:"address"`
	HealthFactor      string `json:"healthFactor"`
	CollateralUsd     string `json:"collateralUsd"`
	DebtUsd           string `json:"debtUsd"`
	LiquidationAmount string `json:"liquidationAmount"`
}

type CollateralBreakdown struct {
	Asset      string  `json:"asset"`
	Amount     string  `json:"amount"`
	ValueUsd   string  `json:"valueUsd"`
	Percentage float64 `json:"percentage"`
}

type TotalCollateral struct {
	Value     string                 `json:"value"`
	Breakdown []CollateralBreakdown  `json:"breakdown"`
}

type StableSupply struct {
	Total       string  `json:"total"`
	Circulating string  `json:"circulating"`
	Backing     float64 `json:"backing"`
}

type ProtocolHealth struct {
	AverageHealthFactor     float64 `json:"averageHealthFactor"`
	UsersAtRisk             int     `json:"usersAtRisk"`
	TotalUsers              int     `json:"totalUsers"`
	CollateralizationRatio  float64 `json:"collateralizationRatio"`
}

type DashboardMetrics struct {
	LiquidatableUsers []LiquidatableUser `json:"liquidatableUsers"`
	TotalCollateral   TotalCollateral    `json:"totalCollateral"`
	StableSupply      StableSupply       `json:"stableSupply"`
	ProtocolHealth    ProtocolHealth     `json:"protocolHealth"`
}
