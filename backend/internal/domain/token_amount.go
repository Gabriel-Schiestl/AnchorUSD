package domain

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model/constants"
)

func GetTokenAmountInUSD(
	amountInWei *big.Int,
	tokenPriceUSD string, 
) (*big.Int, error) {

	priceScaled, ok := ParseDecimalToScaledInt(tokenPriceUSD, constants.PRICE_PRECISION)
	if !ok {
		return nil, fmt.Errorf("invalid token price format")
	}

	usd := new(big.Int).Mul(amountInWei, priceScaled)
	usd.Div(usd, constants.PRECISION)

	return usd, nil
}

func ParseDecimalToScaledInt(value string, scale *big.Int) (*big.Int, bool) {
	parts := strings.Split(value, ".")
	intPart := parts[0]

	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	scaleDigits := len(scale.String()) - 1
	if len(fracPart) > scaleDigits {
		fracPart = fracPart[:scaleDigits]
	}

	for len(fracPart) < scaleDigits {
		fracPart += "0"
	}

	full := intPart + fracPart
	return new(big.Int).SetString(full, 10)
}
