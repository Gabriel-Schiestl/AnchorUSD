package service

import (
	"context"
	"time"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
)

type CollateralHistoryStore interface {
	GetLatestDeposits(ctx context.Context, userAddress string, limit int) ([]model.Deposit, error)
	GetLatestRedeems(ctx context.Context, userAddress string, limit int) ([]model.Redeem, error)
}

type CoinHistoryStore interface {
	GetLatestMints(ctx context.Context, userAddress string, limit int) ([]model.Mints, error)
	GetLatestBurns(ctx context.Context, userAddress string, limit int) ([]model.Burns, error)
}

type LiquidationHistoryStore interface {
	GetLatestLiquidations(ctx context.Context, userAddress string, limit int) ([]model.Liquidations, error)
}

type EventHistoryStore interface {
	GetEventByID(ctx context.Context, eventID uint) (*model.Events, error)
}

type HistoryService struct {
	collateralStore  CollateralHistoryStore
	coinStore        CoinHistoryStore
	liquidationStore LiquidationHistoryStore
	eventsStore      EventHistoryStore
}

func NewHistoryService(collateralStore CollateralHistoryStore, coinStore CoinHistoryStore, liquidationStore LiquidationHistoryStore, eventsStore EventHistoryStore) *HistoryService {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing history service")
	return &HistoryService{
		collateralStore:  collateralStore,
		coinStore:        coinStore,
		liquidationStore: liquidationStore,
		eventsStore:      eventsStore,
	}
}

func (hs *HistoryService) GetUserHistory(ctx context.Context, userAddress string) (model.HistoryData, error) {
	logger := utils.GetLogger()
	logger.Info().Str("user", userAddress).Msg("Fetching user history")

	deposits, err := hs.collateralStore.GetLatestDeposits(ctx, userAddress, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch deposits")
		return model.HistoryData{}, err
	}

	redeems, err := hs.collateralStore.GetLatestRedeems(ctx, userAddress, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch redeems")
		return model.HistoryData{}, err
	}

	mints, err := hs.coinStore.GetLatestMints(ctx, userAddress, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch mints")
		return model.HistoryData{}, err
	}

	burns, err := hs.coinStore.GetLatestBurns(ctx, userAddress, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch burns")
		return model.HistoryData{}, err
	}

	liquidations, err := hs.liquidationStore.GetLatestLiquidations(ctx, userAddress, 10)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch liquidations")
		return model.HistoryData{}, err
	}

	depositTransactions := make([]model.Transaction, 0)
	for _, d := range deposits {
		event, err := hs.eventsStore.GetEventByID(ctx, d.EventID)
		if err != nil || event == nil {
			logger.Warn().Uint("event_id", d.EventID).Msg("Event not found for deposit")
			continue
		}
		depositTransactions = append(depositTransactions, model.Transaction{
			ID:        d.ID,
			Type:      model.TransactionTypeDeposit,
			Amount:    d.Amount.Int.String(),
			Asset:     d.CollateralAddress,
			Timestamp: time.Unix(event.CreatedAt, 0).Format(time.RFC3339),
			TxHash:    event.TxHash,
			Status:    model.TransactionStatusCompleted,
		})
	}

	for _, r := range redeems {
		event, err := hs.eventsStore.GetEventByID(ctx, r.EventID)
		if err != nil || event == nil {
			logger.Warn().Uint("event_id", r.EventID).Msg("Event not found for redeem")
			continue
		}
		depositTransactions = append(depositTransactions, model.Transaction{
			ID:        r.ID,
			Type:      model.TransactionTypeRedeem,
			Amount:    r.Amount.Int.String(),
			Asset:     r.CollateralAddress,
			Timestamp: time.Unix(event.CreatedAt, 0).Format(time.RFC3339),
			TxHash:    event.TxHash,
			Status:    model.TransactionStatusCompleted,
		})
	}

	mintBurnTransactions := make([]model.Transaction, 0)
	for _, m := range mints {
		event, err := hs.eventsStore.GetEventByID(ctx, m.EventID)
		if err != nil || event == nil {
			logger.Warn().Uint("event_id", m.EventID).Msg("Event not found for mint")
			continue
		}
		mintBurnTransactions = append(mintBurnTransactions, model.Transaction{
			ID:        m.ID,
			Type:      model.TransactionTypeMint,
			Amount:    m.Amount.Int.String(),
			Timestamp: time.Unix(event.CreatedAt, 0).Format(time.RFC3339),
			TxHash:    event.TxHash,
			Status:    model.TransactionStatusCompleted,
		})
	}

	for _, b := range burns {
		event, err := hs.eventsStore.GetEventByID(ctx, b.EventID)
		if err != nil || event == nil {
			logger.Warn().Uint("event_id", b.EventID).Msg("Event not found for burn")
			continue
		}
		mintBurnTransactions = append(mintBurnTransactions, model.Transaction{
			ID:        b.ID,
			Type:      model.TransactionTypeBurn,
			Amount:    b.Amount.Int.String(),
			Timestamp: time.Unix(event.CreatedAt, 0).Format(time.RFC3339),
			TxHash:    event.TxHash,
			Status:    model.TransactionStatusCompleted,
		})
	}

	liquidationTransactions := make([]model.Transaction, 0)
	for _, l := range liquidations {
		event, err := hs.eventsStore.GetEventByID(ctx, l.EventID)
		if err != nil || event == nil {
			logger.Warn().Uint("event_id", l.EventID).Msg("Event not found for liquidation")
			continue
		}
		liquidationTransactions = append(liquidationTransactions, model.Transaction{
			ID:        l.ID,
			Type:      model.TransactionTypeLiquidation,
			Amount:    l.DebtCovered.Int.String(),
			Asset:     l.CollateralAddress,
			Timestamp: time.Unix(event.CreatedAt, 0).Format(time.RFC3339),
			TxHash:    event.TxHash,
			Status:    model.TransactionStatusCompleted,
		})
	}

	logger.Info().
		Int("deposits_redeems", len(depositTransactions)).
		Int("mint_burns", len(mintBurnTransactions)).
		Int("liquidations", len(liquidationTransactions)).
		Msg("User history fetched successfully")

	return model.HistoryData{
		Deposits:     depositTransactions,
		MintBurn:     mintBurnTransactions,
		Liquidations: liquidationTransactions,
	}, nil
}
