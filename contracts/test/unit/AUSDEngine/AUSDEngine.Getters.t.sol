// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineMintTest is BaseEngineTest {
    function testGetLiquidationThreshold() external view {
        assertEq(engine.getLiquidationThreshold(), 50);
    }

    function testGetLiquidationPrecision() external view {
        assertEq(engine.getLiquidationPrecision(), 100);
    }

    function testGetMinHealthFactor() external view {
        assertEq(engine.getMinHealthFactor(), 1e18);
    }

    function testGetLiquidationBonus() external view {
        assertEq(engine.getLiquidationBonus(), 10);
    }

    function testGetPrecision() external view {
        assertEq(engine.getPrecision(), 1e18);
    }

    function testGetPriceAdditionalPrecision() external view {
        assertEq(engine.getPriceAdditionalPrecision(), 1e10);
    }

    function testGetTotalCollateralInUSD()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();

        uint256 totalCollateral = engine.getTotalCollateralInUSD(USER);
        // COLLATERAL_DEPOSIT_AMOUNT = 10 ether, ETH price = $2000
        // Expected: 10 * 2000 = $20,000
        assertEq(totalCollateral, 20_000 ether);
    }

    function testGetTotalCollateralInUSDWithMultipleTokens()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.depositCollateral(wbtcAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();

        uint256 totalCollateral = engine.getTotalCollateralInUSD(USER);
        // WETH: 10 * $2000 = $20,000
        // WBTC: 10 * $1000 = $10,000
        // Total: $30,000
        assertEq(totalCollateral, 30_000 ether);
    }

    function testGetTotalCollateralInUSDRevertsWithZeroAddress() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__NotZeroAddress.selector);
        engine.getTotalCollateralInUSD(address(0));
    }

    function testGetHealthFactor() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        uint256 healthFactor = engine.getUserHealthFactor(USER);
        // Collateral: $20,000, Debt: $5
        // Adjusted collateral: $20,000 * 50 / 100 = $10,000
        // Health Factor: $10,000 / $5 = 2000
        assertEq(healthFactor, 2000 ether);
    }

    function testGetHealthFactorWithNoDebt()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        uint256 healthFactor = engine.getHealthFactor();
        assertEq(healthFactor, type(uint256).max);
        vm.stopPrank();
    }

    function testGetCollateralBalanceOfUser()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();

        assertEq(
            engine.getCollateralBalanceOfUser(USER, wethAddr),
            COLLATERAL_DEPOSIT_AMOUNT
        );
    }

    function testGetUserDebt() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), MINT_AMOUNT);
    }

    function testGetUserHealthFactor()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        uint256 healthFactor = engine.getUserHealthFactor(USER);
        assertEq(healthFactor, 2000 ether);
    }

    function testGetUserAccountInformation()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        (uint256 totalCollateral, uint256 debt) = engine
            .getUserAccountInformation(USER);

        assertEq(totalCollateral, 20_000 ether);
        assertEq(debt, MINT_AMOUNT);
    }

    function testGetUserAccountInformationRevertsWithZeroAddress() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__NotZeroAddress.selector);
        engine.getUserAccountInformation(address(0));
    }

    function testGetTokenAmountFromUSD() external view {
        uint256 usdAmount = 2000 ether; // $2000
        uint256 tokenAmount = engine.getTokenAmountFromUSD(wethAddr, usdAmount);

        // ETH price = $2000
        // $2000 / $2000 = 1 ETH
        assertEq(tokenAmount, 1 ether);
    }

    function testGetTokenAmountFromUSDWithBTC() external view {
        uint256 usdAmount = 1000 ether; // $1000
        uint256 tokenAmount = engine.getTokenAmountFromUSD(wbtcAddr, usdAmount);

        // BTC price = $1000
        // $1000 / $1000 = 1 BTC
        assertEq(tokenAmount, 1 ether);
    }

    function testGetTokenAmountFromUSDRevertsWithNotAllowedToken() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__TokenNotAllowed.selector);
        engine.getTokenAmountFromUSD(address(0), 1000 ether);
    }

    function testGetTokenPriceFeed() external view {
        assertEq(engine.getTokenPriceFeed(wethAddr), ethUsdPriceFeed);
        assertEq(engine.getTokenPriceFeed(wbtcAddr), btcUsdPriceFeed);
    }

    function testGetAllowedTokens() external view {
        address[] memory tokens = engine.getAllowedTokens();
        assertEq(tokens.length, 2);
        assertEq(tokens[0], wethAddr);
        assertEq(tokens[1], wbtcAddr);
    }

    function testGetOwner() external view {
        assertEq(engine.getOwner(), deployerAddress);
    }

    function testGetCollateralTokenPrice() external view {
        uint256 ethPrice = engine.getCollateralTokenPrice(wethAddr);
        // ETH price should be $2000 with 18 decimals
        assertEq(ethPrice, 2000 ether);

        uint256 btcPrice = engine.getCollateralTokenPrice(wbtcAddr);
        // BTC price should be $1000 with 18 decimals
        assertEq(btcPrice, 1000 ether);
    }
}
