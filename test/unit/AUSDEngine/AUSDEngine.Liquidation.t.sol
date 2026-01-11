// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineLiquidationTest is BaseEngineTest {
    function testLiquidate() external giveCollateralBalanceAndAllowance {
        // Setup USER with enough collateral to cover liquidation + bonus even after price drop
        vm.startPrank(USER);
        // after price drop, collateral must be higher than debt + 10% bonus
        // debt + 10% bonus = $5.5
        // after price drop, collateral must be >= $5.5
        // if collateral is $20, after 70% price drop it is $6
        uint256 collateralNeeded = engine.getTokenAmountFromUSD(
            wethAddr,
            (MINT_AMOUNT * 4)
        );
        engine.depositCollateral(wethAddr, collateralNeeded);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        // Mock price drop to make USER liquidatable (ETH price drops by 70%)
        int256 newEthPrice = 600e8; // $600 (was $2000)

        vm.mockCall(
            ethUsdPriceFeed,
            abi.encodeWithSignature("latestRoundData()"),
            abi.encode(1, newEthPrice, block.timestamp, block.timestamp, 1)
        );

        // Give liquidator collateral and AUSD to cover debt
        vm.startPrank(deployerAddress);
        weth.transfer(LIQUIDATOR, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(LIQUIDATOR);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        engine.depositCollateral(wethAddr, INITIAL_COLLATERAL_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        // Verify USER health factor is broken
        uint256 userHealthFactorBefore = engine.getUserHealthFactor(USER);
        assertLt(userHealthFactorBefore, engine.getMinHealthFactor());

        // Liquidate
        vm.prank(LIQUIDATOR);
        engine.liquidate(USER, wethAddr, MINT_AMOUNT);

        uint256 tokenAmountLiquidated = engine.getTokenAmountFromUSD(
            wethAddr,
            MINT_AMOUNT
        );
        uint256 bonusCollateral = (tokenAmountLiquidated *
            engine.getLiquidationBonus()) / engine.getLiquidationPrecision();
        uint256 expectedRemaining = collateralNeeded -
            tokenAmountLiquidated -
            bonusCollateral;

        assertEq(engine.getUserDebt(USER), 0);
        assertEq(
            engine.getCollateralBalanceOfUser(USER, wethAddr),
            expectedRemaining
        );

        assertEq(
            weth.balanceOf(LIQUIDATOR),
            tokenAmountLiquidated + bonusCollateral
        );
        assertEq(aUSD.balanceOf(LIQUIDATOR), 0);
    }

    function testIfLiquidateRevertsWhenHealthFactorIsOk()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        // Give liquidator AUSD
        vm.startPrank(deployerAddress);
        weth.transfer(LIQUIDATOR, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(LIQUIDATOR);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        engine.depositCollateral(wethAddr, INITIAL_COLLATERAL_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorOk.selector);
        engine.liquidate(USER, wethAddr, MINT_AMOUNT);
        vm.stopPrank();
    }

    function testIfLiquidateRevertsWhenUserHasNoCollateralOfSpecificToken()
        external
        giveCollateralBalanceAndAllowance
    {
        address USER2 = makeAddr("user2");

        // Setup USER2 with WETH collateral only
        vm.startPrank(deployerAddress);
        weth.transfer(USER2, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(USER2);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(9.5 ether); // Mint very close to maximum
        vm.stopPrank();

        int256 newEthPrice = 180e6;
        vm.mockCall(
            ethUsdPriceFeed,
            abi.encodeWithSignature("latestRoundData()"),
            abi.encode(1, newEthPrice, block.timestamp, block.timestamp, 1)
        );

        // USER2 has no WBTC collateral, trying to liquidate WBTC should fail
        // even though health factor is broken
        vm.expectRevert(AUSDEngine.AUSDEngine__InsufficientCollateral.selector);
        vm.prank(USER);
        engine.liquidate(USER2, wbtcAddr, 1 ether);
    }

    function testLiquidateEmitsEvents()
        external
        giveCollateralBalanceAndAllowance
    {
        // Setup USER with enough collateral
        vm.startPrank(USER);
        uint256 collateralNeeded = engine.getTokenAmountFromUSD(
            wethAddr,
            (MINT_AMOUNT * 4)
        );
        engine.depositCollateral(wethAddr, collateralNeeded);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        // Mock price drop
        int256 newEthPrice = 600e8;
        vm.mockCall(
            ethUsdPriceFeed,
            abi.encodeWithSignature("latestRoundData()"),
            abi.encode(1, newEthPrice, block.timestamp, block.timestamp, 1)
        );

        // Give liquidator AUSD
        vm.startPrank(deployerAddress);
        weth.transfer(LIQUIDATOR, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(LIQUIDATOR);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        engine.depositCollateral(wethAddr, INITIAL_COLLATERAL_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        uint256 tokenAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            MINT_AMOUNT
        );
        uint256 bonusCollateral = ((tokenAmount *
            engine.getLiquidationBonus()) / engine.getLiquidationPrecision());

        // Expect CollateralRedeemed event
        vm.expectEmit(true, true, false, true, address(engine));
        emit CollateralRedeemed(
            USER,
            LIQUIDATOR,
            wethAddr,
            tokenAmount + bonusCollateral
        );

        engine.liquidate(USER, wethAddr, MINT_AMOUNT);
        vm.stopPrank();
    }

    function testLiquidationBonus() external giveCollateralBalanceAndAllowance {
        // Setup USER with enough collateral
        vm.startPrank(USER);
        uint256 collateralNeeded = engine.getTokenAmountFromUSD(
            wethAddr,
            (MINT_AMOUNT * 6)
        );
        engine.depositCollateral(wethAddr, collateralNeeded);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        // Mock price drop
        int256 newEthPrice = 600e8;
        vm.mockCall(
            ethUsdPriceFeed,
            abi.encodeWithSignature("latestRoundData()"),
            abi.encode(1, newEthPrice, block.timestamp, block.timestamp, 1)
        );

        // Setup liquidator
        vm.startPrank(deployerAddress);
        weth.transfer(LIQUIDATOR, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(LIQUIDATOR);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        engine.depositCollateral(wethAddr, INITIAL_COLLATERAL_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        uint256 liquidatorCollateralBefore = engine.getCollateralBalanceOfUser(
            LIQUIDATOR,
            wethAddr
        );

        engine.liquidate(USER, wethAddr, MINT_AMOUNT);

        uint256 liquidatorCollateralAfter = engine.getCollateralBalanceOfUser(
            LIQUIDATOR,
            wethAddr
        );

        // Calculate expected bonus at new price
        uint256 tokenAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            MINT_AMOUNT
        );
        uint256 bonusCollateral = ((tokenAmount *
            engine.getLiquidationBonus()) / engine.getLiquidationPrecision());

        // Liquidator should receive tokenAmount + 10% bonus
        // Use approx equal because of rounding in price conversions
        assertApproxEqAbs(
            liquidatorCollateralAfter,
            liquidatorCollateralBefore + tokenAmount + bonusCollateral,
            1e16 // Allow 0.01 ETH difference for rounding
        );
        vm.stopPrank();
    }
}
