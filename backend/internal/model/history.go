package model

type TransactionType string

const (
	TransactionTypeDeposit      TransactionType = "deposit"
	TransactionTypeMint         TransactionType = "mint"
	TransactionTypeBurn         TransactionType = "burn"
	TransactionTypeLiquidation  TransactionType = "liquidation"
	TransactionTypeRedeem       TransactionType = "redeem"
)

type TransactionStatus string

const (
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID        string            `json:"id"`
	Type      TransactionType   `json:"type"`
	Amount    string            `json:"amount"`
	Asset     string            `json:"asset,omitempty"`
	Timestamp string            `json:"timestamp"`
	TxHash    string            `json:"txHash"`
	Status    TransactionStatus `json:"status"`
}

type HistoryData struct {
	Deposits     []Transaction `json:"deposits"`
	MintBurn     []Transaction `json:"mintBurn"`
	Liquidations []Transaction `json:"liquidations"`
}
