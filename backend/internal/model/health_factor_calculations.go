package model

type CalculateMintRequest struct {
	Address    string `json:"address" binding:"required"`
	MintAmount string `json:"mintAmount" binding:"required"`
}

type CalculateBurnRequest struct {
	Address    string `json:"address" binding:"required"`
	BurnAmount string `json:"burnAmount" binding:"required"`
}

type CalculateDepositRequest struct {
	Address       string `json:"address" binding:"required"`
	TokenAddress  string `json:"tokenAddress" binding:"required"`
	DepositAmount string `json:"depositAmount" binding:"required"`
}

type HealthFactorProjection struct {
	HealthFactorAfter  string `json:"healthFactorAfter"`
	NewDebt            string `json:"newDebt"`
	NewCollateralValue string `json:"newCollateralValue"`
}
