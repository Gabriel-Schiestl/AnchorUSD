//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Script} from "forge-std/Script.sol";

contract HelperConfig is Script {
    
    struct NetworkConfig {
        address ethUsdPriceFeed;
    }

    NetworkConfig activeConfig;

    function getSepoliaConfig() internal {

    }

    function getAnvilConfig() internal {

    }

    function getActiveConfig() external returns(NetworkConfig) {
        if(activeConfig.ethUsdPriceFeed != address(0)) {
            return activeConfig;
        }

        if(block.chainId == 11155111) {
            activeConfig = getSepoliaConfig();
        } else {
            activeConfig = getAnvilConfig();
        }

        return activeConfig;
    }
}