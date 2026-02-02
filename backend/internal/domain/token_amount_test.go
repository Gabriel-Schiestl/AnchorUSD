package domain

import (
	"math/big"
	"testing"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func TestGetTokenAmountInUSD(t *testing.T) {
	tests := []struct {
		name          string
		amountInWei   *big.Int
		tokenPriceUSD string
		shouldError   bool
		checkPositive bool
	}{
		{
			name:          "1 ETH at $2000",
			amountInWei:   constants.PRECISION,                            // 1 ETH
			tokenPriceUSD: "2000.00",
			shouldError:   false,
			checkPositive: true,
		},
		{
			name:          "0.5 ETH at $2000",
			amountInWei:   new(big.Int).Div(constants.PRECISION, big.NewInt(2)), // 0.5 ETH
			tokenPriceUSD: "2000.00",
			shouldError:   false,
			checkPositive: true,
		},
		{
			name:          "1 BTC at $40000",
			amountInWei:   constants.PRECISION,
			tokenPriceUSD: "40000.00",
			shouldError:   false,
			checkPositive: true,
		},
		{
			name:          "zero amount",
			amountInWei:   big.NewInt(0),
			tokenPriceUSD: "2000.00",
			shouldError:   false,
			checkPositive: false,
		},
		{
			name:          "invalid price format",
			amountInWei:   constants.PRECISION,
			tokenPriceUSD: "invalid",
			shouldError:   true,
			checkPositive: false,
		},
		{
			name:          "price with many decimals",
			amountInWei:   constants.PRECISION,
			tokenPriceUSD: "2000.123456789",
			shouldError:   false,
			checkPositive: true,
		},
		{
			name:          "very small amount",
			amountInWei:   big.NewInt(1), // 1 wei
			tokenPriceUSD: "2000.00",
			shouldError:   false,
			checkPositive: false, // Will be very small, possibly zero after division
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetTokenAmountInUSD(tt.amountInWei, tt.tokenPriceUSD)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Result should not be nil")
				return
			}

			if tt.checkPositive && result.Sign() <= 0 {
				t.Errorf("Expected positive result, got %s", result.String())
			}
		})
	}
}

func TestGetTokenAmountInUSDCalculation(t *testing.T) {
	amountInWei := constants.PRECISION // 1 ETH (1e18 wei)
	tokenPriceUSD := "2000.00"         // $2000 per ETH

	result, err := GetTokenAmountInUSD(amountInWei, tokenPriceUSD)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Expected: 1 ETH * $2000 = $2000
	// With PRICE_PRECISION (1e8), price becomes 200000000000
	// Formula: (amountInWei * priceScaled) / PRECISION
	// = (1e18 * 200000000000) / 1e18 = 200000000000
	
	expectedStr := "200000000000" // $2000 with 8 decimals precision
	if result.String() != expectedStr {
		t.Errorf("Calculation mismatch. Got %s, want %s", result.String(), expectedStr)
	}
}

func TestGetTokenAmountInUSDFormula(t *testing.T) {
	amountInWei := new(big.Int).Mul(big.NewInt(5), constants.PRECISION) // 5 tokens
	tokenPriceUSD := "100.00"                                            // $100 per token

	result, err := GetTokenAmountInUSD(amountInWei, tokenPriceUSD)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	priceScaled, ok := ParseDecimalToScaledInt(tokenPriceUSD, constants.PRICE_PRECISION)
	if !ok {
		t.Fatal("Failed to parse price")
	}

	expected := new(big.Int).Mul(amountInWei, priceScaled)
	expected.Div(expected, constants.PRECISION)

	if result.Cmp(expected) != 0 {
		t.Errorf("Formula verification failed. Got %s, want %s", result.String(), expected.String())
	}
}

func TestParseDecimalToScaledInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		scale    *big.Int
		expected string
		ok       bool
	}{
		{
			name:     "whole number",
			value:    "2000",
			scale:    constants.PRICE_PRECISION, // 1e8
			expected: "200000000000",
			ok:       true,
		},
		{
			name:     "number with decimals",
			value:    "2000.50",
			scale:    constants.PRICE_PRECISION,
			expected: "200050000000",
			ok:       true,
		},
		{
			name:     "number with many decimals",
			value:    "2000.12345678",
			scale:    constants.PRICE_PRECISION,
			expected: "200012345678",
			ok:       true,
		},
		{
			name:     "number with excess decimals (truncated)",
			value:    "2000.123456789012",
			scale:    constants.PRICE_PRECISION,
			expected: "200012345678", // Truncated to 8 decimals
			ok:       true,
		},
		{
			name:     "zero",
			value:    "0",
			scale:    constants.PRICE_PRECISION,
			expected: "0",
			ok:       true,
		},
		{
			name:     "zero with decimals",
			value:    "0.00",
			scale:    constants.PRICE_PRECISION,
			expected: "0",
			ok:       true,
		},
		{
			name:     "small decimal",
			value:    "0.01",
			scale:    constants.PRICE_PRECISION,
			expected: "1000000",
			ok:       true,
		},
		{
			name:     "number with fewer decimals than scale",
			value:    "100.5",
			scale:    constants.PRICE_PRECISION,
			expected: "10050000000",
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseDecimalToScaledInt(tt.value, tt.scale)

			if ok != tt.ok {
				t.Errorf("ParseDecimalToScaledInt() ok = %v, want %v", ok, tt.ok)
				return
			}

			if !tt.ok {
				return
			}

			if result == nil {
				t.Error("Result should not be nil when ok is true")
				return
			}

			if result.String() != tt.expected {
				t.Errorf("ParseDecimalToScaledInt() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestParseDecimalToScaledIntPadding(t *testing.T) {
	value := "100.5"
	scale := big.NewInt(1e8) // 8 decimal places

	result, ok := ParseDecimalToScaledInt(value, scale)
	if !ok {
		t.Fatal("Failed to parse")
	}

	expected := "10050000000"
	if result.String() != expected {
		t.Errorf("Padding failed. Got %s, want %s", result.String(), expected)
	}
}

func TestParseDecimalToScaledIntTruncation(t *testing.T) {
	value := "100.123456789" // 9 decimal places
	scale := big.NewInt(1e8)  // 8 decimal places

	result, ok := ParseDecimalToScaledInt(value, scale)
	if !ok {
		t.Fatal("Failed to parse")
	}

	expected := "10012345678"
	if result.String() != expected {
		t.Errorf("Truncation failed. Got %s, want %s", result.String(), expected)
	}
}

func TestParseDecimalToScaledIntNoDecimals(t *testing.T) {
	value := "2000"
	scale := constants.PRICE_PRECISION

	result, ok := ParseDecimalToScaledInt(value, scale)
	if !ok {
		t.Fatal("Failed to parse")
	}

	expected := "200000000000"
	if result.String() != expected {
		t.Errorf("Whole number parsing failed. Got %s, want %s", result.String(), expected)
	}
}

func TestGetTokenAmountInUSDEdgeCases(t *testing.T) {
	t.Run("very large amount", func(t *testing.T) {
		largeAmount := new(big.Int).Mul(big.NewInt(1000000), constants.PRECISION)
		result, err := GetTokenAmountInUSD(largeAmount, "2000.00")
		
		if err != nil {
			t.Errorf("Should handle large amounts: %v", err)
		}
		
		if result.Sign() <= 0 {
			t.Error("Large amount should produce positive result")
		}
	})

	t.Run("empty price string", func(t *testing.T) {
		result, err := GetTokenAmountInUSD(constants.PRECISION, "")
		
		if err != nil {
			if result != nil {
				t.Error("Result should be nil on error")
			}
		} else {
			if result.Sign() != 0 {
				t.Error("Empty price should produce zero result")
			}
		}
	})
}
