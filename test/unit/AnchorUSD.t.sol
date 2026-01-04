//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Test} from "forge-std/Test.sol";
import {AnchorUSD} from "../../src/AnchorUSD.sol";
import {DeployAUSD} from "../../script/DeployAUSD.s.sol";

contract AnchorUSDTest is Test {
    AnchorUSD aUSD;
    DeployAUSD deployer;

    function setUp() external {
        deployer = new DeployAUSD();
        aUSD = deployer.run();
    }
}
