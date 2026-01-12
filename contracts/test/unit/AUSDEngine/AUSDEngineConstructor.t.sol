// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Test, console} from "forge-std/Test.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";
import {AnchorUSD} from "../../../src/AnchorUSD.sol";
import {DeployAUSD} from "../../../script/DeployAUSD.s.sol";
import {HelperConfig} from "../../../script/HelperConfig.s.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

//////// Constructor Tests ////////

contract AUSDEngineConstructorTest is Test {
    address wethAddr = makeAddr("weth");
    address wbtcAddr = makeAddr("wbtc");
    address ethUsdPriceFeed = makeAddr("ethUsdPriceFeed");
    address btcUsdPriceFeed = makeAddr("btcUsdPriceFeed");

    function testConstructorRevertsWithMismatchedArrays() external {
        address[] memory tokens = new address[](2);
        tokens[0] = wethAddr;
        tokens[1] = wbtcAddr;

        address[] memory priceFeeds = new address[](1);
        priceFeeds[0] = ethUsdPriceFeed;

        vm.expectRevert(
            AUSDEngine
                .AUSDEngine__TokenAddressesAndPriceFeedAddressesAmountsDontMatch
                .selector
        );
        new AUSDEngine(tokens, priceFeeds);
    }

    function testConstructorRevertsWithZeroTokenAddress() external {
        address[] memory tokens = new address[](2);
        tokens[0] = address(0);
        tokens[1] = wbtcAddr;

        address[] memory priceFeeds = new address[](2);
        priceFeeds[0] = ethUsdPriceFeed;
        priceFeeds[1] = btcUsdPriceFeed;

        vm.expectRevert(AUSDEngine.AUSDEngine__NotZeroAddress.selector);
        new AUSDEngine(tokens, priceFeeds);
    }

    function testConstructorRevertsWithZeroPriceFeedAddress() external {
        address[] memory tokens = new address[](2);
        tokens[0] = wethAddr;
        tokens[1] = wbtcAddr;

        address[] memory priceFeeds = new address[](2);
        priceFeeds[0] = address(0);
        priceFeeds[1] = btcUsdPriceFeed;

        vm.expectRevert(AUSDEngine.AUSDEngine__NotZeroAddress.selector);
        new AUSDEngine(tokens, priceFeeds);
    }

    function testConstructorSetsOwner() external {
        address[] memory tokens = new address[](1);
        tokens[0] = wethAddr;

        address[] memory priceFeeds = new address[](1);
        priceFeeds[0] = ethUsdPriceFeed;

        AUSDEngine engine = new AUSDEngine(tokens, priceFeeds);

        assertEq(engine.getOwner(), address(this));
    }

    function testConstructorSetsTokensAndPriceFeeds() external {
        address[] memory tokens = new address[](2);
        tokens[0] = wethAddr;
        tokens[1] = wbtcAddr;

        address[] memory priceFeeds = new address[](2);
        priceFeeds[0] = ethUsdPriceFeed;
        priceFeeds[1] = btcUsdPriceFeed;

        AUSDEngine engine = new AUSDEngine(tokens, priceFeeds);

        assertEq(engine.getTokenPriceFeed(wethAddr), ethUsdPriceFeed);
        assertEq(engine.getTokenPriceFeed(wbtcAddr), btcUsdPriceFeed);

        address[] memory allowedTokens = engine.getAllowedTokens();
        assertEq(allowedTokens.length, 2);
        assertEq(allowedTokens[0], wethAddr);
        assertEq(allowedTokens[1], wbtcAddr);
    }
}
