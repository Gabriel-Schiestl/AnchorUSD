// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineRedeemTest is BaseEngineTest {
    function testRedeemCollateral() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.redeemCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 0);
        assertEq(weth.balanceOf(address(engine)), 0);
        assertEq(weth.balanceOf(USER), INITIAL_COLLATERAL_AMOUNT);
    }

    function testIfRedeemCollateralEmitsEvent()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        vm.expectEmit(true, true, true, false, address(engine));
        emit CollateralRedeemed(USER, wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        engine.redeemCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();
    }

    function testIfRedeemCollateralRevertsWithNoCollateralBalance() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__InsufficientCollateral.selector);
        vm.prank(USER);
        engine.redeemCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
    }

    function testIfRedeemCollateralRevertsWhenHealthFactorBreaks()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorBroken.selector);
        engine.redeemCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();
    }

    function testRedeemCollateralForAUSD()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        engine.redeemCollateralForAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT,
            MINT_AMOUNT
        );
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), 0);
        assertEq(aUSD.balanceOf(USER), 0);
        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 0);
    }

    function testIfRedeemCollateralForAUSDRevertsWhenHealthFactorBreaks()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorBroken.selector);
        engine.redeemCollateralForAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT,
            MINT_AMOUNT - 1 ether
        );
        vm.stopPrank();
    }

    function testIfRedeemCollateralForAUSDRevertsWithInsufficientCollateral()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__InsufficientCollateral.selector);
        engine.redeemCollateralForAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT + 1 ether,
            MINT_AMOUNT - 1 ether
        );
        vm.stopPrank();
    }

    function testIfRedeemCollateralForAUSDRevertsWithAmountHigherThanDebt()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__BurnAmountExceedsDebt.selector);
        engine.redeemCollateralForAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT,
            MINT_AMOUNT + 1 ether
        );
        vm.stopPrank();
    }

    function testPartialRedemption()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.redeemCollateral(wethAddr, 3 ether);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 7 ether);
    }

    function testRedeemDifferentTokenThanDeposited()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, 5 ether);
        engine.depositCollateral(wbtcAddr, 3 ether);

        // Can redeem WBTC even though WETH was deposited first
        engine.redeemCollateral(wbtcAddr, 1 ether);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 5 ether);
        assertEq(engine.getCollateralBalanceOfUser(USER, wbtcAddr), 2 ether);
    }
}
