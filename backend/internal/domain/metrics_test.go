package domain

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func TestCalculateLiquidationAmount(t *testing.T) {
	tests := []struct {
		name     string
		debt     *big.Int
		expected *big.Int
	}{
		{
			name:     "nil debt returns zero",
			debt:     nil,
			expected: big.NewInt(0),
		},
		{
			name:     "zero debt returns zero",
			debt:     big.NewInt(0),
			expected: big.NewInt(0),
		},
		{
			name:     "100 debt returns 50",
			debt:     new(big.Int).Mul(big.NewInt(100), constants.PRECISION),
			expected: new(big.Int).Mul(big.NewInt(50), constants.PRECISION),
		},
		{
			name:     "1000 debt returns 500",
			debt:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION),
			expected: new(big.Int).Mul(big.NewInt(500), constants.PRECISION),
		},
		{
			name:     "odd number debt",
			debt:     new(big.Int).Mul(big.NewInt(101), constants.PRECISION),
			expected: new(big.Int).Div(new(big.Int).Mul(big.NewInt(101), constants.PRECISION), big.NewInt(2)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLiquidationAmount(tt.debt)
			
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("CalculateLiquidationAmount() = %s, want %s", result.String(), tt.expected.String())
			}
		})
	}
}

func TestCalculateLiquidationAmountFormula(t *testing.T) {
	debt := new(big.Int).Mul(big.NewInt(1000), constants.PRECISION)
	result := CalculateLiquidationAmount(debt)
	
	expected := new(big.Int).Div(debt, constants.LIQUIDATION_DIVISOR)
	
	if result.Cmp(expected) != 0 {
		t.Errorf("Liquidation amount formula incorrect. Got %s, want %s", result.String(), expected.String())
	}
}

func TestPercentageOf(t *testing.T) {
	tests := []struct {
		name      string
		value     *big.Int
		total     *big.Int
		expected  float64
		tolerance float64
	}{
		{
			name:      "50% of total",
			value:     big.NewInt(50),
			total:     big.NewInt(100),
			expected:  50.0,
			tolerance: 0.01,
		},
		{
			name:      "25% of total",
			value:     big.NewInt(25),
			total:     big.NewInt(100),
			expected:  25.0,
			tolerance: 0.01,
		},
		{
			name:      "100% of total",
			value:     big.NewInt(100),
			total:     big.NewInt(100),
			expected:  100.0,
			tolerance: 0.01,
		},
		{
			name:      "0% of total",
			value:     big.NewInt(0),
			total:     big.NewInt(100),
			expected:  0.0,
			tolerance: 0.0,
		},
		{
			name:      "more than 100%",
			value:     big.NewInt(150),
			total:     big.NewInt(100),
			expected:  150.0,
			tolerance: 0.01,
		},
		{
			name:      "nil total returns 0",
			value:     big.NewInt(50),
			total:     nil,
			expected:  0.0,
			tolerance: 0.0,
		},
		{
			name:      "zero total returns 0",
			value:     big.NewInt(50),
			total:     big.NewInt(0),
			expected:  0.0,
			tolerance: 0.0,
		},
		{
			name:      "decimal percentage",
			value:     big.NewInt(1),
			total:     big.NewInt(3),
			expected:  33.33,
			tolerance: 0.5, // Allow some tolerance for integer division
		},
		{
			name:      "very small percentage",
			value:     big.NewInt(1),
			total:     big.NewInt(1000),
			expected:  0.1,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PercentageOf(tt.value, tt.total)
			
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("PercentageOf(%s, %s) = %v, want %v (tolerance %v)",
					tt.value.String(), 
					func() string {
						if tt.total == nil {
							return "nil"
						}
						return tt.total.String()
					}(),
					result, 
					tt.expected, 
					tt.tolerance)
			}
		})
	}
}

func TestPercentageOfWithLargeNumbers(t *testing.T) {
	value := new(big.Int).Mul(big.NewInt(1000000), constants.PRECISION)
	total := new(big.Int).Mul(big.NewInt(2000000), constants.PRECISION)
	
	result := PercentageOf(value, total)
	expected := 50.0
	
	if result < expected-0.01 || result > expected+0.01 {
		t.Errorf("PercentageOf with large numbers = %v, want ~%v", result, expected)
	}
}

func TestPercentageOfFormula(t *testing.T) {
	value := big.NewInt(50)
	total := big.NewInt(100)
	
	result := PercentageOf(value, total)
	
	// Manual calculation
	pctBig := new(big.Int).Mul(value, constants.PERCENTAGE_MULTIPLIER)
	pctBig.Div(pctBig, total)
	expected := float64(pctBig.Int64()) / constants.PERCENTAGE_BASE_DIVISOR
	
	if result != expected {
		t.Errorf("Formula verification failed. Got %v, want %v", result, expected)
	}
}
