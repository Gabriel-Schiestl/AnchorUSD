// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.30;

import { Script } from "forge-std/Script.sol";
import { HelperConfig } from "./HelperConfig.s.sol";
import { AnchorUSD } from "../src/AnchorUSD.sol";
import { AUSDEngine } from "../src/AUSDEngine.sol";

contract DeployAUSD is Script {
    address[] public tokenAddresses;
    address[] public priceFeedAddresses;

    function run() external returns (AnchorUSD, AUSDEngine, HelperConfig) {
        HelperConfig helperConfig = new HelperConfig();

        (address weth, address wbtc, address ethUsdPriceFeed, address btcUsdPriceFeed, uint256 deployerKey) =
            helperConfig.activeNetworkConfig();
        tokenAddresses = [weth, wbtc];
        priceFeedAddresses = [ethUsdPriceFeed, btcUsdPriceFeed];

        vm.startBroadcast(deployerKey);
        AUSDEngine aUSDEngine = new AUSDEngine(tokenAddresses, priceFeedAddresses);
        AnchorUSD aUSD = new AnchorUSD(address(aUSDEngine));
        aUSDEngine.setAUSD(aUSD);
        vm.stopBroadcast();
        return (aUSD, aUSDEngine, helperConfig);
    }
}