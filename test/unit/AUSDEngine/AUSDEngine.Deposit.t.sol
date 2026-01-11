// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineDepositTest is BaseEngineTest {
    function testDepositCollateral()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.prank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        assertEq(
            engine.getCollateralBalanceOfUser(USER, wethAddr),
            COLLATERAL_DEPOSIT_AMOUNT
        );
        assertEq(weth.balanceOf(address(engine)), COLLATERAL_DEPOSIT_AMOUNT);
        assertEq(
            weth.balanceOf(USER),
            INITIAL_COLLATERAL_AMOUNT - COLLATERAL_DEPOSIT_AMOUNT
        );
    }

    function testIfDepositCollateralEmitsEvent()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.expectEmit(true, true, true, false, address(engine));
        emit CollateralDeposited(USER, wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        vm.prank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
    }

    function testDepositCollateralAndMintAUSD()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateralAndMintAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT,
            MINT_AMOUNT
        );
        vm.stopPrank();

        assertEq(
            engine.getCollateralBalanceOfUser(USER, wethAddr),
            COLLATERAL_DEPOSIT_AMOUNT
        );
        assertEq(engine.getUserDebt(USER), MINT_AMOUNT);
        assertEq(aUSD.balanceOf(USER), MINT_AMOUNT);
    }

    function testIfDepositCollateralAndMintAUSDRevertsWithZeroCollateral()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.expectRevert(AUSDEngine.AUSDEngine__MustBeMoreThanZero.selector);
        vm.prank(USER);
        engine.depositCollateralAndMintAUSD(wethAddr, 0, MINT_AMOUNT);
    }

    function testIfDepositCollateralAndMintAUSDRevertsWithZeroMintAmount()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.expectRevert(AUSDEngine.AUSDEngine__MustBeMoreThanZero.selector);
        vm.prank(USER);
        engine.depositCollateralAndMintAUSD(
            wethAddr,
            COLLATERAL_DEPOSIT_AMOUNT,
            0
        );
    }

    function testIfDepositCollateralAndMintAUSDRevertsWithNotAllowedToken()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.expectRevert(AUSDEngine.AUSDEngine__TokenNotAllowed.selector);
        vm.prank(USER);
        engine.depositCollateralAndMintAUSD(
            address(0),
            COLLATERAL_DEPOSIT_AMOUNT,
            MINT_AMOUNT
        );
    }

    function testIfDepositCollateralAndMintAUSDRevertsWhenHealthFactorBreaks()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        uint256 tokenAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            (MINT_AMOUNT * 2)
        );

        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorBroken.selector);
        engine.depositCollateralAndMintAUSD(
            wethAddr,
            tokenAmount,
            MINT_AMOUNT + 1 ether
        );
        vm.stopPrank();
    }

    function testDepositMultipleCollateralTypes()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, 5 ether);
        engine.depositCollateral(wbtcAddr, 3 ether);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 5 ether);
        assertEq(engine.getCollateralBalanceOfUser(USER, wbtcAddr), 3 ether);
    }

    function testMultipleDepositsToSameToken()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, 5 ether);
        engine.depositCollateral(wethAddr, 3 ether);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 8 ether);
    }

    function testHealthFactorAtExactLimit()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        // Deposit $200 worth of collateral
        uint256 collateralAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            200 ether
        );
        engine.depositCollateral(wethAddr, collateralAmount);

        // Mint $100 AUSD (exactly at 200% collateralization)
        engine.mintAUSD(100 ether);

        uint256 healthFactor = engine.getHealthFactor();
        // Health factor should be exactly 1e18 (at the limit)
        assertEq(healthFactor, 1e18);
        vm.stopPrank();
    }
}
