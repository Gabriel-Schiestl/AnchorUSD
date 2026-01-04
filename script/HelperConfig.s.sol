//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Script} from "forge-std/Script.sol";

contract HelperConfig is Script {
    struct NetworkConfig {
        address ethUsdPriceFeed;
    }

    NetworkConfig activeConfig;

    function getSepoliaConfig() internal returns (NetworkConfig memory) {}

    function getAnvilConfig() internal returns (NetworkConfig memory) {}

    function getActiveConfig() external returns (NetworkConfig memory) {
        if (activeConfig.ethUsdPriceFeed != address(0)) {
            return activeConfig;
        }

        if (block.chainid == 11155111) {
            activeConfig = getSepoliaConfig();
        } else {
            activeConfig = getAnvilConfig();
        }

        return activeConfig;
    }
}
