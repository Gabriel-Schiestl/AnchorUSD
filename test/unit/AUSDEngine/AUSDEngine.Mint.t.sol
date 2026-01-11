// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {BaseEngineTest} from "./BaseEngine.t.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";

contract AUSDEngineMintTest is BaseEngineTest {
    function testMintAUSD() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), MINT_AMOUNT);
        assertEq(aUSD.balanceOf(USER), MINT_AMOUNT);
    }

    function testIfMintAUSDEmitsEvent()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);

        vm.expectEmit(true, true, false, false, address(engine));
        emit AUSDMinted(USER, MINT_AMOUNT);

        engine.mintAUSD(MINT_AMOUNT);
        vm.stopPrank();
    }

    function testIfMintAUSDRevertsWhenHealthFactorBreaks()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        // $10 => 10e18 wei
        // $ price => 2000e18
        // 10e18 * 1e18 / 2000e18 => 5e15
        // 5e15 wei => 0.005 ether
        uint256 tokenAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            (MINT_AMOUNT * 2)
        );
        engine.depositCollateral(wethAddr, tokenAmount);

        // collateral = $10, mint = $6 => health factor < 1
        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorBroken.selector);
        engine.mintAUSD(MINT_AMOUNT + 1 ether);
        vm.stopPrank();
    }

    function testMultipleMints() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.mintAUSD(2 ether);
        engine.mintAUSD(3 ether);
        vm.stopPrank();

        assertEq(engine.getUserDebt(USER), 5 ether);
        assertEq(aUSD.balanceOf(USER), 5 ether);
    }

    function testCannotMintBelowHealthFactorLimit()
        external
        giveCollateralBalanceAndAllowance
    {
        vm.startPrank(USER);
        uint256 collateralAmount = engine.getTokenAmountFromUSD(
            wethAddr,
            200 ether
        );
        engine.depositCollateral(wethAddr, collateralAmount);

        // Try to mint $100.01 AUSD (would break health factor)
        vm.expectRevert(AUSDEngine.AUSDEngine__HealthFactorBroken.selector);
        engine.mintAUSD(100 ether + 0.01 ether);
        vm.stopPrank();
    }
}
