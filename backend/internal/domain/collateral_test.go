package domain

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func TestCollateralizationRatio(t *testing.T) {
	tests := []struct {
		name            string
		totalCollateral *big.Int
		totalDebt       *big.Int
		expected        float64
	}{
		{
			name:            "200% collateralization",
			totalCollateral: big.NewInt(200),
			totalDebt:       big.NewInt(100),
			expected:        200.0,
		},
		{
			name:            "150% collateralization",
			totalCollateral: big.NewInt(150),
			totalDebt:       big.NewInt(100),
			expected:        150.0,
		},
		{
			name:            "zero debt returns 0",
			totalCollateral: big.NewInt(1000),
			totalDebt:       big.NewInt(0),
			expected:        0.0,
		},
		{
			name:            "nil debt returns 0",
			totalCollateral: big.NewInt(1000),
			totalDebt:       nil,
			expected:        0.0,
		},
		{
			name:            "100% collateralization",
			totalCollateral: big.NewInt(100),
			totalDebt:       big.NewInt(100),
			expected:        100.0,
		},
		{
			name:            "under-collateralized 50%",
			totalCollateral: big.NewInt(50),
			totalDebt:       big.NewInt(100),
			expected:        50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CollateralizationRatio(tt.totalCollateral, tt.totalDebt)
			if result != tt.expected {
				t.Errorf("CollateralizationRatio() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateBackingPercentage(t *testing.T) {
	tests := []struct {
		name            string
		totalCollateral *big.Int
		circulating     *big.Int
		expected        float64
	}{
		{
			name:            "200% backing",
			totalCollateral: big.NewInt(2000),
			circulating:     big.NewInt(1000),
			expected:        200.0,
		},
		{
			name:            "100% backing",
			totalCollateral: big.NewInt(1000),
			circulating:     big.NewInt(1000),
			expected:        100.0,
		},
		{
			name:            "50% backing (under-backed)",
			totalCollateral: big.NewInt(500),
			circulating:     big.NewInt(1000),
			expected:        50.0,
		},
		{
			name:            "zero circulating returns 0",
			totalCollateral: big.NewInt(1000),
			circulating:     big.NewInt(0),
			expected:        0.0,
		},
		{
			name:            "nil circulating returns 0",
			totalCollateral: big.NewInt(1000),
			circulating:     nil,
			expected:        0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBackingPercentage(tt.totalCollateral, tt.circulating)
			if result != tt.expected {
				t.Errorf("CalculateBackingPercentage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateCollateralDeposited(t *testing.T) {
	tests := []struct {
		name     string
		assets   []CollateralAssetData
		expected int
	}{
		{
			name: "multiple assets with valid amounts",
			assets: []CollateralAssetData{
				{
					Name:     "ETH",
					Amount:   new(big.Int).Mul(big.NewInt(1), constants.PRECISION), // 1 ETH
					PriceUSD: "2000.00",
				},
				{
					Name:     "BTC",
					Amount:   new(big.Int).Mul(big.NewInt(1), constants.PRECISION), // 1 BTC
					PriceUSD: "40000.00",
				},
			},
			expected: 2,
		},
		{
			name: "skip zero amounts",
			assets: []CollateralAssetData{
				{
					Name:     "ETH",
					Amount:   big.NewInt(0),
					PriceUSD: "2000.00",
				},
				{
					Name:     "BTC",
					Amount:   new(big.Int).Mul(big.NewInt(1), constants.PRECISION),
					PriceUSD: "40000.00",
				},
			},
			expected: 1,
		},
		{
			name: "skip nil amounts",
			assets: []CollateralAssetData{
				{
					Name:     "ETH",
					Amount:   nil,
					PriceUSD: "2000.00",
				},
				{
					Name:     "BTC",
					Amount:   new(big.Int).Mul(big.NewInt(1), constants.PRECISION),
					PriceUSD: "40000.00",
				},
			},
			expected: 1,
		},
		{
			name:     "empty assets",
			assets:   []CollateralAssetData{},
			expected: 0,
		},
		{
			name: "all zero amounts",
			assets: []CollateralAssetData{
				{
					Name:     "ETH",
					Amount:   big.NewInt(0),
					PriceUSD: "2000.00",
				},
				{
					Name:     "BTC",
					Amount:   big.NewInt(0),
					PriceUSD: "40000.00",
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCollateralDeposited(tt.assets)
			if len(result) != tt.expected {
				t.Errorf("CalculateCollateralDeposited() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestCalculateCollateralDepositedValues(t *testing.T) {
	assets := []CollateralAssetData{
		{
			Name:     "ETH",
			Amount:   new(big.Int).Mul(big.NewInt(1), constants.PRECISION), // 1 ETH
			PriceUSD: "2000.00",
		},
	}

	result := CalculateCollateralDeposited(assets)
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0].Asset != "ETH" {
		t.Errorf("Expected asset name 'ETH', got '%s'", result[0].Asset)
	}

	expectedAmount := new(big.Int).Mul(big.NewInt(1), constants.PRECISION).String()
	if result[0].Amount != expectedAmount {
		t.Errorf("Expected amount '%s', got '%s'", expectedAmount, result[0].Amount)
	}

	if result[0].ValueUsd == "" {
		t.Error("ValueUsd should not be empty")
	}
}
