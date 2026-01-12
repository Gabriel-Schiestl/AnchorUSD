//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Test, console} from "forge-std/Test.sol";
import {AnchorUSD} from "../../../src/AnchorUSD.sol";
import {AUSDEngine} from "../../../src/AUSDEngine.sol";
import {DeployAUSD} from "../../../script/DeployAUSD.s.sol";
import {ERC20Mock} from "../../mocks/ERC20Mock.sol";
import {MockV3Aggregator} from "@chainlink/contracts/src/v0.8/tests/MockV3Aggregator.sol";

contract FailOnRevertHandler is Test {
    AnchorUSD aUSD;
    AUSDEngine engine;

    ERC20Mock public weth;
    ERC20Mock public wbtc;

    uint96 private constant MAX_DEPOSIT = type(uint96).max;

    constructor(AnchorUSD _aUSD, AUSDEngine _engine) {
        aUSD = _aUSD;
        engine = _engine;

        address[] memory collateralTokens = engine.getAllowedTokens();
        weth = ERC20Mock(collateralTokens[0]);
        wbtc = ERC20Mock(collateralTokens[1]);
    }

    function mintAndDepositCollateral(
        uint256 collateralIndex,
        uint256 collateralAmount
    ) external {
        bound(collateralIndex, 1, type(uint256).max);
        collateralAmount = bound(collateralAmount, 1, MAX_DEPOSIT);

        ERC20Mock collateral = _getCollateralToken(collateralIndex);

        vm.startPrank(msg.sender);
        collateral.mint(msg.sender, collateralAmount);
        collateral.approve(address(engine), collateralAmount);
        engine.depositCollateral(address(collateral), collateralAmount);
        vm.stopPrank();
    }

    function redeemCollateral(
        uint256 collateralIndex,
        uint256 collateralAmount
    ) external {
        bound(collateralIndex, 1, type(uint256).max);
        ERC20Mock collateral = _getCollateralToken(collateralIndex);

        collateralAmount = bound(
            collateralAmount,
            0,
            engine.getCollateralBalanceOfUser(msg.sender, address(collateral))
        );
        if (collateralAmount == 0) return;

        vm.prank(msg.sender);
        engine.redeemCollateral(address(collateral), collateralAmount);
    }

    function burnAUSD(uint256 aUSDAmount) public {
        // Must burn more than 0
        aUSDAmount = bound(aUSDAmount, 0, aUSD.balanceOf(msg.sender));
        if (aUSDAmount == 0) {
            return;
        }
        vm.startPrank(msg.sender);
        aUSD.approve(address(engine), aUSDAmount);
        engine.burnAUSD(aUSDAmount);
        vm.stopPrank();
    }

    /////// AUSD ///////

    function transferAUSD(uint256 amountAUSD, address to) public {
        if (to == address(0)) {
            to = address(1);
        }
        amountAUSD = bound(amountAUSD, 0, aUSD.balanceOf(msg.sender));
        vm.prank(msg.sender);
        aUSD.transfer(to, amountAUSD);
    }

    /////// Aggregator ///////

    function updateCollateralPrice(
        uint96 newPrice,
        uint256 collateralIndex
    ) public {
        bound(collateralIndex, 1, type(uint256).max);
        newPrice = uint96(bound(uint256(newPrice), 1, type(uint96).max));
        int256 intNewPrice = int256(uint256(newPrice));

        ERC20Mock collateral = _getCollateralToken(collateralIndex);
        MockV3Aggregator priceFeed = MockV3Aggregator(
            engine.getTokenPriceFeed(address(collateral))
        );

        priceFeed.updateAnswer(intNewPrice);
    }

    function _getCollateralToken(uint256 index) private returns (ERC20Mock) {
        if (index % 2 == 0) return weth;

        return wbtc;
    }
}
