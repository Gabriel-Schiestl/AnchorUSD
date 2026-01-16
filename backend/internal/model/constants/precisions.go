package constants

import "math/big"

var LIQUIDATION_THRESHOLD = big.NewInt(50)
var LIQUIDATION_PRECISION = big.NewInt(100)
var PRECISION = big.NewInt(1e18)
var MIN_HEALTH_FACTOR = big.NewInt(1e18)
var MAX_HEALTH_FACTOR = big.NewInt(9e18)
var PRICE_PRECISION = big.NewInt(1e8)