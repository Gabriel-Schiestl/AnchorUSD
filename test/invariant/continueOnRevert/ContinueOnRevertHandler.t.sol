//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Test, console} from "forge-std/Test.sol";
import {AnchorUSD} from "../../../src/AnchorUSD.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";
import {DeployAUSD} from "../../../script/DeployAUSD.s.sol";
import {ERC20Mock} from "../../mocks/ERC20Mock.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";

contract ContinueOnRevertHandler is Test {
    AnchorUSD aUSD;
    AUSDEngine engine;

    AggregatorV3Interface public ethUsdPriceFeed;
    AggregatorV3Interface public btcUsdPriceFeed;
    ERC20Mock public weth;
    ERC20Mock public wbtc;

    uint96 private constant MAX_DEPOSIT = type(uint96).max;

    constructor(AnchorUSD _aUSD, AUSDEngine _engine) {
        aUSD = _aUSD;
        engine = _engine;

        address[] memory collateralTokens = engine.getAllowedTokens();
        weth = ERC20Mock(collateralTokens[0]);
        wbtc = ERC20Mock(collateralTokens[1]);

        ethUsdPriceFeed = AggregatorV3Interface(
            engine.getTokenPriceFeed(address(weth))
        );
        btcUsdPriceFeed = AggregatorV3Interface(
            engine.getTokenPriceFeed(address(wbtc))
        );
    }

    function mintAndDepositCollateral(
        uint256 collateralIndex,
        uint256 collateralAmount
    ) external {
        collateralAmount = bound(collateralAmount, 0, MAX_DEPOSIT);

        ERC20Mock collateral = _getCollateralToken(collateralIndex);

        collateral.mint(msg.sender, collateralAmount);
        collateral.approve(address(engine), collateralAmount);
        engine.depositCollateral(address(collateral), collateralAmount);
    }

    function redeemCollateral(
        uint256 collateralIndex,
        uint256 collateralAmount
    ) external {
        ERC20Mock collateral = _getCollateralToken(collateralIndex);

        engine.redeemCollateral(address(collateral), collateralAmount);
    }

    function burnAUSD(uint256 aUSDAmount) public {
        aUSDAmount = bound(aUSDAmount, 0, aUSD.balanceOf(msg.sender));
        engine.burnAUSD(aUSDAmount);
    }

    function mintAUSD(uint256 amountAUSD) public {
        amountAUSD = bound(amountAUSD, 0, MAX_DEPOSIT);
        engine.mintAUSD(amountAUSD);
    }

    function liquidate(
        uint256 collateralSeed,
        address userToBeLiquidated,
        uint256 debtToCover
    ) public {
        ERC20Mock collateral = _getCollateralToken(collateralSeed);
        engine.liquidate(address(collateral), userToBeLiquidated, debtToCover);
    }

    function _getCollateralToken(uint256 index) private returns (ERC20Mock) {
        if (index % 2 == 0) return weth;

        return wbtc;
    }

    function callSummary() external view {
        console.log("Weth total deposited", weth.balanceOf(address(engine)));
        console.log("Wbtc total deposited", wbtc.balanceOf(address(engine)));
        console.log("Total supply of DSC", aUSD.totalSupply());
    }
}
