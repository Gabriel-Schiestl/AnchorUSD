// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineBurnTest is BaseEngineTest {
    function testBurnAUSD() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        engine.burnAUSD(MINT_AMOUNT);
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), 0);
        assertEq(aUSD.balanceOf(USER), 0);
    }

    function testIfBurnAUSDEmitsEvent()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectEmit(true, true, false, false, address(engine));
        emit AUSDBurned(USER, MINT_AMOUNT);

        engine.burnAUSD(MINT_AMOUNT);
        vm.stopPrank();
    }

    function testIfBurnAUSDRevertsWithAmountHigherThanDebt()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);

        vm.expectRevert(AUSDEngine.AUSDEngine__BurnAmountExceedsDebt.selector);
        engine.burnAUSD(MINT_AMOUNT + 1 ether);
        vm.stopPrank();
    }

    function testPartialBurn() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        engine.burnAUSD(2 ether);
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), 3 ether);
        assertEq(aUSD.balanceOf(USER), 3 ether);
    }
}
