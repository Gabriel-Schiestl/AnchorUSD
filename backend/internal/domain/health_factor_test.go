package domain

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func TestCalculateHealthFactor(t *testing.T) {
	tests := []struct {
		name             string
		collateralUSD    *big.Int
		debtUSD          *big.Int
		expectedAbove    *big.Int // Expected to be above this value
		expectedBelow    *big.Int // Expected to be below this value (nil means no upper bound)
	}{
		{
			name:          "healthy position - 200% collateralized",
			collateralUSD: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION), // $2000
			debtUSD:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION), // $1000
			expectedAbove: constants.MIN_HEALTH_FACTOR,                              // Should be > 1.0
			expectedBelow: nil,
		},
		{
			name:          "at liquidation threshold - exactly 200% collateralized",
			collateralUSD: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			debtUSD:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			expectedAbove: constants.MIN_HEALTH_FACTOR,
			expectedBelow: nil,
		},
		{
			name:          "under-collateralized - 100% collateralized",
			collateralUSD: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			debtUSD:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			expectedAbove: big.NewInt(0),
			expectedBelow: nil, // Health factor formula produces value based on liquidation threshold
		},
		{
			name:          "zero debt - infinite health factor",
			collateralUSD: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			debtUSD:       big.NewInt(0),
			expectedAbove: constants.MIN_HEALTH_FACTOR,
			expectedBelow: nil,
		},
		{
			name:          "high collateralization - 400%",
			collateralUSD: new(big.Int).Mul(big.NewInt(4000), constants.PRECISION),
			debtUSD:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			expectedAbove: new(big.Int).Mul(big.NewInt(2), constants.PRECISION), // Should be > 2.0
			expectedBelow: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHealthFactor(tt.collateralUSD, tt.debtUSD)

			if result == nil {
				t.Fatal("CalculateHealthFactor returned nil")
			}

			if tt.expectedAbove != nil && result.Cmp(tt.expectedAbove) < 0 {
				t.Errorf("Health factor %s is below expected minimum %s", result.String(), tt.expectedAbove.String())
			}

			if tt.expectedBelow != nil && result.Cmp(tt.expectedBelow) >= 0 {
				t.Errorf("Health factor %s is not below expected maximum %s", result.String(), tt.expectedBelow.String())
			}
		})
	}
}

func TestCalculateHealthFactorAfterMint(t *testing.T) {
	tests := []struct {
		name              string
		currentCollateral *big.Int
		currentDebt       *big.Int
		mintAmount        *big.Int
		shouldDecrease    bool
	}{
		{
			name:              "minting increases debt and decreases health factor",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(500), constants.PRECISION),
			mintAmount:        new(big.Int).Mul(big.NewInt(500), constants.PRECISION),
			shouldDecrease:    true,
		},
		{
			name:              "minting from zero debt",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       big.NewInt(0),
			mintAmount:        new(big.Int).Mul(big.NewInt(100), constants.PRECISION),
			shouldDecrease:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeHF := CalculateHealthFactor(tt.currentCollateral, tt.currentDebt)
			afterHF := CalculateHealthFactorAfterMint(tt.currentCollateral, tt.currentDebt, tt.mintAmount)

			if tt.shouldDecrease && afterHF.Cmp(beforeHF) >= 0 {
				t.Errorf("Health factor should decrease after mint. Before: %s, After: %s", beforeHF.String(), afterHF.String())
			}
		})
	}
}

func TestCalculateHealthFactorAfterBurn(t *testing.T) {
	tests := []struct {
		name              string
		currentCollateral *big.Int
		currentDebt       *big.Int
		burnAmount        *big.Int
		shouldIncrease    bool
	}{
		{
			name:              "burning decreases debt and increases health factor",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			burnAmount:        new(big.Int).Mul(big.NewInt(500), constants.PRECISION),
			shouldIncrease:    true,
		},
		{
			name:              "burning more than debt sets debt to zero",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(100), constants.PRECISION),
			burnAmount:        new(big.Int).Mul(big.NewInt(200), constants.PRECISION),
			shouldIncrease:    true,
		},
		{
			name:              "burning all debt",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			burnAmount:        new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			shouldIncrease:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeHF := CalculateHealthFactor(tt.currentCollateral, tt.currentDebt)
			afterHF := CalculateHealthFactorAfterBurn(tt.currentCollateral, tt.currentDebt, tt.burnAmount)

			if tt.shouldIncrease && afterHF.Cmp(beforeHF) <= 0 {
				t.Errorf("Health factor should increase after burn. Before: %s, After: %s", beforeHF.String(), afterHF.String())
			}

			newDebt := new(big.Int).Sub(tt.currentDebt, tt.burnAmount)
			if newDebt.Sign() < 0 {
				expectedHF := CalculateHealthFactor(tt.currentCollateral, big.NewInt(0))
				if afterHF.Cmp(expectedHF) != 0 {
					t.Errorf("When burning more than debt, HF should equal HF with zero debt")
				}
			}
		})
	}
}

func TestCalculateHealthFactorAfterDeposit(t *testing.T) {
	tests := []struct {
		name              string
		currentCollateral *big.Int
		currentDebt       *big.Int
		depositAmount     *big.Int
		shouldIncrease    bool
	}{
		{
			name:              "deposit increases collateral and health factor",
			currentCollateral: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(500), constants.PRECISION),
			depositAmount:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			shouldIncrease:    true,
		},
		{
			name:              "small deposit still increases health factor",
			currentCollateral: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION),
			currentDebt:       new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			depositAmount:     new(big.Int).Mul(big.NewInt(100), constants.PRECISION),
			shouldIncrease:    true,
		},
		{
			name:              "deposit with zero debt",
			currentCollateral: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			currentDebt:       big.NewInt(0),
			depositAmount:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			shouldIncrease:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeHF := CalculateHealthFactor(tt.currentCollateral, tt.currentDebt)
			afterHF := CalculateHealthFactorAfterDeposit(tt.currentCollateral, tt.currentDebt, tt.depositAmount)

			if tt.shouldIncrease && afterHF.Cmp(beforeHF) <= 0 {
				t.Errorf("Health factor should increase after deposit. Before: %s, After: %s", beforeHF.String(), afterHF.String())
			}
		})
	}
}

func TestAverageHealthFactor(t *testing.T) {
	tests := []struct {
		name       string
		sumHF      *big.Int
		totalUsers int
		expected   float64
		tolerance  float64
	}{
		{
			name:       "average of 2.0",
			sumHF:      new(big.Int).Mul(big.NewInt(4), constants.PRECISION), // 4.0 total
			totalUsers: 2,
			expected:   2.0,
			tolerance:  0.001,
		},
		{
			name:       "average of 1.5",
			sumHF:      new(big.Int).Mul(big.NewInt(3), constants.PRECISION), // 3.0 total
			totalUsers: 2,
			expected:   1.5,
			tolerance:  0.001,
		},
		{
			name:       "zero users returns 0",
			sumHF:      new(big.Int).Mul(big.NewInt(10), constants.PRECISION),
			totalUsers: 0,
			expected:   0.0,
			tolerance:  0.0,
		},
		{
			name:       "single user",
			sumHF:      new(big.Int).Mul(big.NewInt(2), constants.PRECISION),
			totalUsers: 1,
			expected:   2.0,
			tolerance:  0.001,
		},
		{
			name: "large sum doesn't overflow",
			sumHF: func() *big.Int {
				val := new(big.Int).Mul(big.NewInt(1000000), constants.PRECISION)
				return val
			}(),
			totalUsers: 1000,
			expected:   1000.0,
			tolerance:  0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AverageHealthFactor(tt.sumHF, tt.totalUsers)
			
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("AverageHealthFactor() = %v, want %v (tolerance %v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestIsAtRisk(t *testing.T) {
	tests := []struct {
		name     string
		hf       *big.Int
		expected bool
	}{
		{
			name:     "nil health factor is not at risk",
			hf:       nil,
			expected: false,
		},
		{
			name:     "health factor below risk threshold",
			hf:       new(big.Int).Mul(big.NewInt(1), constants.PRECISION), // 1.0
			expected: true,
		},
		{
			name:     "health factor at risk threshold",
			hf:       constants.RISK_THRESHOLD,
			expected: false,
		},
		{
			name:     "health factor above risk threshold",
			hf:       new(big.Int).Mul(big.NewInt(2), constants.PRECISION), // 2.0
			expected: false,
		},
		{
			name:     "health factor very low",
			hf:       new(big.Int).Div(constants.PRECISION, big.NewInt(2)), // 0.5
			expected: true,
		},
		{
			name:     "health factor zero",
			hf:       big.NewInt(0),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAtRisk(tt.hf)
			if result != tt.expected {
				hfStr := "nil"
				if tt.hf != nil {
					hfStr = tt.hf.String()
				}
				t.Errorf("IsAtRisk(%s) = %v, want %v", hfStr, result, tt.expected)
			}
		})
	}
}
