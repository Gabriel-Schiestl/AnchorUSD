// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Test, console} from "forge-std/Test.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";
import {AnchorUSD} from "../../../src/AnchorUSD.sol";
import {DeployAUSD} from "../../../script/DeployAUSD.s.sol";
import {HelperConfig} from "../../../script/HelperConfig.s.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract BaseEngineTest is Test {
    AnchorUSD aUSD;
    AUSDEngine engine;
    DeployAUSD deployer;
    HelperConfig config;

    address LIQUIDATOR = makeAddr("liquidator");

    address internal USER = makeAddr("user");
    uint256 internal constant INITIAL_COLLATERAL_AMOUNT = 10 ether;
    uint256 internal constant COLLATERAL_DEPOSIT_AMOUNT = 10 ether;
    uint256 internal constant MINT_AMOUNT = 5 ether;

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
    event CollateralRedeemed(
        address indexed user,
        address indexed token,
        uint256 indexed amount
    );
    event AUSDMinted(address indexed user, uint256 indexed amount);
    event AUSDBurned(address indexed user, uint256 indexed amount);
    event Liquidation(
        address indexed liquidatedUser,
        address indexed liquidator,
        address indexed tokenCollateral,
        uint256 collateralAmount,
        uint256 debtCovered
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

    function testOnlyOwnerCanSetAUSD() external {
        address notOwner = makeAddr("notOwner");

        // Deploy new engine
        address[] memory tokens = new address[](1);
        tokens[0] = wethAddr;
        address[] memory priceFeeds = new address[](1);
        priceFeeds[0] = ethUsdPriceFeed;

        AUSDEngine newEngine = new AUSDEngine(tokens, priceFeeds);

        vm.expectRevert(AUSDEngine.AUSDEngine__NotOwner.selector);
        vm.prank(notOwner);
        newEngine.setAUSD(aUSD);
    }
}
