// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Test, console} from "forge-std/Test.sol";
import {AUSDEngine} from "../../src/AUSDEngine.sol";
import {AnchorUSD} from "../../src/AnchorUSD.sol";
import {DeployAUSD} from "../../script/DeployAUSD.s.sol";
import {HelperConfig} from "../../script/HelperConfig.s.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract AUSDEngineTest is Test {
    AnchorUSD aUSD;
    AUSDEngine engine;
    DeployAUSD deployer;
    HelperConfig config;

    address private USER = makeAddr("user");
    uint256 private constant INITIAL_COLLATERAL_AMOUNT = 10;
    uint256 private constant COLLATERAL_DEPOSIT_AMOUNT = 10;
    uint256 private constant MINT_AMOUNT = 5;

    address wethAddr;
    address wbtcAddr;
    IERC20 weth;
    IERC20 wbtc;
    address ethUsdPriceFeed;
    address btcUsdPriceFeed;
    uint256 deployerKey;

    address deployerAddress;

    event CollateralDeposited(
        address indexed user,
        address indexed token,
        uint256 indexed amount
    );

    function setUp() external {
        deployer = new DeployAUSD();
        (aUSD, engine, config) = deployer.run();

        (
            wethAddr,
            wbtcAddr,
            ethUsdPriceFeed,
            btcUsdPriceFeed,
            deployerKey
        ) = config.activeNetworkConfig();

        deployerAddress = vm.addr(deployerKey);
        weth = IERC20(wethAddr);
        wbtc = IERC20(wbtcAddr);
    }

    modifier giveCollateralBalanceAndAllowance() {
        vm.startPrank(deployerAddress);
        weth.transfer(USER, INITIAL_COLLATERAL_AMOUNT);
        wbtc.transfer(USER, INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();

        vm.startPrank(USER);
        weth.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        wbtc.approve(address(engine), INITIAL_COLLATERAL_AMOUNT);
        vm.stopPrank();
        _;
    }

    function testRevertIfAmountIsZero() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__MustBeMoreThanZero.selector);
        engine.depositCollateral(wethAddr, 0);
    }

    function testRevertIfNotAllowedToken() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__TokenNotAllowed.selector);
        engine.depositCollateral(address(0), INITIAL_COLLATERAL_AMOUNT);
    }

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

    function testRedeemCollateral() external giveCollateralBalanceAndAllowance {
        vm.startPrank(USER);
        engine.depositCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        engine.redeemCollateral(wethAddr, COLLATERAL_DEPOSIT_AMOUNT);
        vm.stopPrank();

        assertEq(engine.getCollateralBalanceOfUser(USER, wethAddr), 0);
        assertEq(weth.balanceOf(address(engine)), 0);
        assertEq(weth.balanceOf(USER), INITIAL_COLLATERAL_AMOUNT);
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

    function testRevertSetAUSDIfNotOwner() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__NotOwner.selector);
        engine.setAUSD(aUSD);
    }

    function testRevertIfAUSDIsSet() external {
        vm.expectRevert(AUSDEngine.AUSDEngine__AUSDAlreadyDefined.selector);
        vm.prank(deployerAddress);
        engine.setAUSD(aUSD);
    }
}
