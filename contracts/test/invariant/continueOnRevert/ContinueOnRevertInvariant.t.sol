//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {StdInvariant} from "forge-std/StdInvariant.sol";
import {Test, console} from "forge-std/Test.sol";
import {AnchorUSD} from "../../../src/AnchorUSD.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";
import {DeployAUSD} from "../../../script/DeployAUSD.s.sol";
import {ContinueOnRevertHandler} from "./ContinueOnRevertHandler.t.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {HelperConfig} from "../../../script/HelperConfig.s.sol";

contract ContinueOnRevertInvariantTest is StdInvariant, Test {
    AnchorUSD aUSD;
    AUSDEngine engine;
    HelperConfig config;

    address weth;
    address wbtc;

    function setUp() external {
        DeployAUSD deployer = new DeployAUSD();
        (aUSD, engine, config) = deployer.run();
        (weth, wbtc, , , ) = config.activeNetworkConfig();
        ContinueOnRevertHandler handler = new ContinueOnRevertHandler(
            aUSD,
            engine
        );
        targetContract(address(handler));
    }

    /// forge-config: default.invariant.fail-on-revert = false
    function invariant_protocolMustHaveMoreValueThatTotalSupplyDollars()
        public
        view
    {
        uint256 totalSupply = aUSD.totalSupply();
        uint256 wethDeposited = IERC20(weth).balanceOf(address(engine));
        uint256 wbtcDeposited = IERC20(wbtc).balanceOf(address(engine));

        uint256 wethPrice = engine.getCollateralTokenPrice(weth);
        uint256 wbtcPrice = engine.getCollateralTokenPrice(wbtc);

        uint256 wethValue = wethDeposited * wethPrice;
        uint256 wbtcValue = wbtcDeposited * wbtcPrice;

        console.log("wethValue: %s", wethValue);
        console.log("wbtcValue: %s", wbtcValue);

        assert(wethValue + wbtcValue >= totalSupply);
    }
}
