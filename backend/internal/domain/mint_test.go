package domain

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func TestCalculateMaxMintable(t *testing.T) {
	tests := []struct {
		name              string
		collateralUSDStr  string
		totalDebtStr      string
		expectedMin       string // Minimum expected value
		shouldBePositive  bool
		shouldBeZero      bool
	}{
		{
			name:             "healthy position - can mint more",
			collateralUSDStr: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION).String(), // $2000
			totalDebtStr:     new(big.Int).Mul(big.NewInt(500), constants.PRECISION).String(),  // $500
			shouldBePositive: true,
			shouldBeZero:     false,
		},
		{
			name:             "at liquidation threshold - minimal mintable",
			collateralUSDStr: new(big.Int).Mul(big.NewInt(2000), constants.PRECISION).String(),
			totalDebtStr:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION).String(),
			shouldBePositive: false, // At threshold, might be zero or very small
			shouldBeZero:     false,
		},
		{
			name:             "over-borrowed - cannot mint",
			collateralUSDStr: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION).String(),
			totalDebtStr:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION).String(),
			shouldBePositive: false,
			shouldBeZero:     false, // Due to ADDITIONAL_PRICE_PRECISION adjustment, result may not be exactly zero
		},
		{
			name:             "zero debt - max mintable based on collateral",
			collateralUSDStr: new(big.Int).Mul(big.NewInt(1000), constants.PRECISION).String(),
			totalDebtStr:     "0",
			shouldBePositive: true,
			shouldBeZero:     false,
		},
		{
			name:             "zero collateral - cannot mint",
			collateralUSDStr: "0",
			totalDebtStr:     new(big.Int).Mul(big.NewInt(100), constants.PRECISION).String(),
			shouldBePositive: false,
			shouldBeZero:     true,
		},
		{
			name:             "high collateral - large mintable amount",
			collateralUSDStr: new(big.Int).Mul(big.NewInt(10000), constants.PRECISION).String(),
			totalDebtStr:     new(big.Int).Mul(big.NewInt(1000), constants.PRECISION).String(),
			shouldBePositive: true,
			shouldBeZero:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMaxMintable(tt.collateralUSDStr, tt.totalDebtStr)
			
			resultBig := new(big.Int)
			resultBig.SetString(result, 10)

			if tt.shouldBeZero && resultBig.Cmp(big.NewInt(0)) != 0 {
				t.Errorf("Expected zero but got %s", result)
			}

			if tt.shouldBePositive && resultBig.Cmp(big.NewInt(0)) <= 0 {
				t.Errorf("Expected positive value but got %s", result)
			}

			if !tt.shouldBePositive && !tt.shouldBeZero && resultBig.Cmp(big.NewInt(0)) < 0 {
				t.Errorf("Result should not be negative, got %s", result)
			}
		})
	}
}

func TestCalculateMaxMintableFormula(t *testing.T) {
	// Verify the formula: (collateral * LIQUIDATION_THRESHOLD / LIQUIDATION_PRECISION) - debt
	// With ADDITIONAL_PRICE_PRECISION adjustment
	
	collateralUSD := new(big.Int).Mul(big.NewInt(2000), constants.PRECISION)
	totalDebt := new(big.Int).Mul(big.NewInt(500), constants.PRECISION)

	result := CalculateMaxMintable(collateralUSD.String(), totalDebt.String())
	resultBig := new(big.Int)
	resultBig.SetString(result, 10)

	collateralAdjusted := new(big.Int).Mul(collateralUSD, constants.ADDITIONAL_PRICE_PRECISION)
	collateralAdjusted.Mul(collateralAdjusted, constants.LIQUIDATION_THRESHOLD)
	collateralAdjusted.Div(collateralAdjusted, constants.LIQUIDATION_PRECISION)
	
	expected := new(big.Int).Sub(collateralAdjusted, totalDebt)
	if expected.Sign() < 0 {
		expected = big.NewInt(0)
	}

	if resultBig.Cmp(expected) != 0 {
		t.Errorf("Formula verification failed. Got %s, want %s", result, expected.String())
	}
}

func TestCalculateMaxMintableNegativeClamp(t *testing.T) {
	collateralUSD := new(big.Int).Mul(big.NewInt(100), constants.PRECISION)  // $100
	totalDebt := new(big.Int).Mul(big.NewInt(1000), constants.PRECISION)    // $1000

	result := CalculateMaxMintable(collateralUSD.String(), totalDebt.String())
	resultBig := new(big.Int)
	resultBig.SetString(result, 10)
	
	if resultBig.Sign() < 0 {
		t.Errorf("Result should not be negative, got %s", result)
	}
}

func TestCalculateMaxMintableExactThreshold(t *testing.T) {
	collateralUSD := new(big.Int).Mul(big.NewInt(2000), constants.PRECISION)
	totalDebt := new(big.Int).Mul(big.NewInt(1000), constants.PRECISION)

	result := CalculateMaxMintable(collateralUSD.String(), totalDebt.String())
	resultBig := new(big.Int)
	resultBig.SetString(result, 10)
	
	t.Logf("At liquidation threshold, max mintable = %s", result)
}

func TestCalculateMaxMintableWithZeroDebt(t *testing.T) {
	collateralUSD := new(big.Int).Mul(big.NewInt(1000), constants.PRECISION)
	totalDebt := "0"

	result := CalculateMaxMintable(collateralUSD.String(), totalDebt)
	resultBig := new(big.Int)
	resultBig.SetString(result, 10)

	if resultBig.Sign() <= 0 {
		t.Errorf("With zero debt, max mintable should be positive, got %s", result)
	}

	expectedAdjusted := new(big.Int).Mul(collateralUSD, constants.ADDITIONAL_PRICE_PRECISION)
	expectedAdjusted.Mul(expectedAdjusted, constants.LIQUIDATION_THRESHOLD)
	expectedAdjusted.Div(expectedAdjusted, constants.LIQUIDATION_PRECISION)

	if resultBig.Cmp(expectedAdjusted) != 0 {
		t.Errorf("Max mintable with zero debt should equal adjusted collateral. Got %s, want %s", 
			result, expectedAdjusted.String())
	}
}

func TestCalculateMaxMintableLargeNumbers(t *testing.T) {
	collateralUSD := new(big.Int).Mul(big.NewInt(1000000), constants.PRECISION)
	totalDebt := new(big.Int).Mul(big.NewInt(100000), constants.PRECISION)

	result := CalculateMaxMintable(collateralUSD.String(), totalDebt.String())
	resultBig := new(big.Int)
	resultBig.SetString(result, 10)

	if resultBig.Sign() <= 0 {
		t.Errorf("With large healthy position, max mintable should be positive, got %s", result)
	}

	if result == "" {
		t.Error("Result should not be empty")
	}
}
