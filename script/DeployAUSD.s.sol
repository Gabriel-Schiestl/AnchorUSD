//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Script} from "forge-std/Script.sol";
import {HelperConfig} from "./HelperConfig.s.sol";

contract DeployAUSD is Script {
    HelperConfig config;

    function run() external {
        config = new HelperConfig(); 
    }
}