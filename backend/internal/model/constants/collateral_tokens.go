package constants

import (
	"os"
	"strings"
)

var CollateralTokens = map[string]string{}

func LoadCollateralTokens() {
	collaterals := os.Getenv("COLLATERAL_TOKEN_ADDRESSES")
	collateralsNames := os.Getenv("COLLATERAL_TOKEN_NAMES")

	addresses := strings.Split(collaterals, ",")

	for i, name := range strings.Split(collateralsNames, ",") {
		CollateralTokens[name] = addresses[i]
	}
}
