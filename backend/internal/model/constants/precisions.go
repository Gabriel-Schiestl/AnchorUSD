package constants

import "math/big"

var LIQUIDATION_THRESHOLD = big.NewInt(50)
var LIQUIDATION_PRECISION = big.NewInt(100)
var PRECISION = big.NewInt(1e18)
var MIN_HEALTH_FACTOR = big.NewInt(1e18)
var MAX_HEALTH_FACTOR = big.NewInt(9e18)
var PRICE_PRECISION = big.NewInt(1e8)
var LIQUIDATION_DIVISOR = big.NewInt(2)
var PERCENTAGE_MULTIPLIER = big.NewInt(10000)
var PERCENTAGE_BASE_DIVISOR = 100.0
var RISK_THRESHOLD = new(big.Int).Mul(big.NewInt(15), big.NewInt(1e17))
