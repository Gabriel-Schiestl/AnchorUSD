//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {AnchorUSD} from "./AnchorUSD.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";

contract AUSDEngine {
    error AUSDEngine__NotOwner();
    error AUSDEngine__AUSDAlreadyDefined();
    error AUSDEngine__NotZeroAddress();
    error AUSDEngine__InvalidPrice();

    uint256 private constant PRICE_ADITIONAL_PRECISION = 1e10;
    uint256 private constant PRECISION = 1e18;

    AnchorUSD private s_ausd;
    address private immutable i_owner;
    mapping(address user => mapping(address token => uint256 collateral)) private s_collateralDeposited;
    mapping(address user => uint256 debt) private s_totalDept;
    mapping(address token => address priceFeed) private s_priceFeeds;
    address[] private tokensAllowed;

    constructor(address weth, address wbtc, address wethPriceFeed, address wbtcPriceFeed) {
        i_owner = msg.sender;
        tokensAllowed.push(weth);
        tokensAllowed.push(wbtc);
        s_priceFeeds[weth] = wethPriceFeed;
        s_priceFeeds[wbtc] = wbtcPriceFeed;
    }

    modifier onlyOwner() {
        if(msg.sender != i_owner) {
            revert AUSDEngine__NotOwner();
        }
        _;
    }

    function depositCollateral(address token, uint256 _amout) public {}

    function redeemCollateral(address token, uint256 _amout) public {}

    function redeemCollateralForAUSD(address token, uint256 amount) public {}

    function mintAUSD(uint256 _amout) public {}

    function depositCollateralAndMintAUSD(address token, uint256 collateralAmount, uint256 aUSDAmount) public {}

    function getAccountInformation(address user) public view returns(uint256 totalUsdCollateral, uint256 aUSDDebt) {
        if(user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        aUSDDebt = s_totalDept[user];
        totalUsdCollateral = getTotalCollateralInUSD(user);
    }

    function getTotalCollateralInUSD(address user) public view returns(uint256 totalAmount) {
        if(user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        address[] memory tokens = tokensAllowed;
        for(uint256 i = 0; i < tokens.length; i++) {
            totalAmount += _getCollateralInUSD(tokens[i], user);
        }
    }

    function _getCollateralInUSD(address token, address user) private view returns(uint256) {
        AggregatorV3Interface priceFeed = AggregatorV3Interface(s_priceFeeds[token]);

        (,int256 price,,,) = priceFeed.latestRoundData();

        if(price < 0) {
            revert AUSDEngine__InvalidPrice();
        }

        uint256 priceAdjusted = uint256(price) * PRICE_ADITIONAL_PRECISION;

        uint256 collateral = s_collateralDeposited[user][token];

        uint256 usdCollateral = (priceAdjusted * collateral) / PRECISION;

        return usdCollateral;
    }

    function setAUSD(AnchorUSD _ausd) public onlyOwner {
        if(address(s_ausd) != address(0)) {
            revert AUSDEngine__AUSDAlreadyDefined();
        }

        s_ausd = _ausd;
    }
}