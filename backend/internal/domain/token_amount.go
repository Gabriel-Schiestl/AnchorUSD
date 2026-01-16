package domain

import "math/big"

func GetTokenAmountInUSD(amountInWei *big.Int, tokenPriceUSD float64) *big.Int {
	tokenPriceUSDBigInt := big.NewInt(0).SetInt64(int64(tokenPriceUSD * 1e18))

	amountInUSD := big.NewInt(0).Mul(amountInWei, tokenPriceUSDBigInt)
	amountInUSD.Div(amountInUSD, PRECISION)
	
	return amountInUSD
}