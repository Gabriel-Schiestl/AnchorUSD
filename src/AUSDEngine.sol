//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {AnchorUSD} from "./AnchorUSD.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

contract AUSDEngine {
    using SafeERC20 for IERC20;

    error AUSDEngine__NotOwner();
    error AUSDEngine__AUSDAlreadyDefined();
    error AUSDEngine__NotZeroAddress();
    error AUSDEngine__InvalidPrice();
    error AUSDEngine__TokenNotAllowed();
    error AUSDEngine__MustBeMoreThanZero();
    error AUSDEngine__InsufficientCollateral();
    error AUSDEngine__HealthFactorBroken();

    uint256 private constant PRICE_ADITIONAL_PRECISION = 1e10;
    uint256 private constant PRECISION = 1e18;
    uint256 private constant LIQUIDATION_THRESHOLD = 50;
    uint256 private constant LIQUIDATION_PRECISION = 100;
    uint256 private constant MIN_HEALTH_FACTOR = 1e18;

    AnchorUSD private s_ausd;
    address private immutable i_owner;
    mapping(address user => mapping(address token => uint256 collateral)) private s_collateralDeposited;
    mapping(address user => uint256 debt) private s_totalDept;
    mapping(address token => address priceFeed) private s_priceFeeds;
    address[] private s_tokensAllowed;

    constructor(address weth, address wbtc, address wethPriceFeed, address wbtcPriceFeed) {
        i_owner = msg.sender;
        s_tokensAllowed.push(weth);
        s_tokensAllowed.push(wbtc);
        s_priceFeeds[weth] = wethPriceFeed;
        s_priceFeeds[wbtc] = wbtcPriceFeed;
    }

    event CollateralDeposited(address indexed user, address indexed token, uint256 indexed amount);
    event CollateralRedeemed(address indexed user, address indexed token, uint256 indexed amount);
    event AUSDMinted(address indexed user, uint256 indexed amount);

    modifier onlyOwner() {
        if(msg.sender != i_owner) {
            revert AUSDEngine__NotOwner();
        }
        _;
    }

    modifier onlyAllowedTokens(address token) {
        if(s_priceFeeds[token] == address(0)) {
            revert AUSDEngine__TokenNotAllowed();
        }
        _;
    }

    modifier moreThanZero(uint256 amount) {
        if(amount == 0) {
            revert AUSDEngine__MustBeMoreThanZero();
        }
        _;
    }

    function depositCollateral(address token, uint256 _amount) public onlyAllowedTokens(token) moreThanZero(_amount) {
        if(token == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        s_collateralDeposited[msg.sender][token] += _amount;
        emit CollateralDeposited(msg.sender, token, _amount);

        IERC20(token).safeTransferFrom(msg.sender, address(this), _amount);
    }

    function redeemCollateral(address token, uint256 _amount) public onlyAllowedTokens(token) moreThanZero(_amount) {
        uint256 collateralDeposited = s_collateralDeposited[msg.sender][token];
        if(_amount > collateralDeposited) {
            revert AUSDEngine__InsufficientCollateral();
        }

        s_collateralDeposited[msg.sender][token] -= _amount;
        emit CollateralRedeemed(msg.sender, token, _amount);

        _revertIfHealthFactorBroken(msg.sender);

        IERC20(token).safeTransfer(msg.sender, _amount);
    }

    function redeemCollateralForAUSD(address token, uint256 _collateralAmount, uint256 aUSDToBurn) public onlyAllowedTokens(token) moreThanZero(_collateralAmount) moreThanZero(aUSDToBurn) {
        _burnAUSD(msg.sender, msg.sender, aUSDToBurn);
        redeemCollateral(token, _collateralAmount);
    }

    function mintAUSD(uint256 _amount) public moreThanZero(_amount) {
        s_totalDept[msg.sender] += _amount;
        emit AUSDMinted(msg.sender, _amount);

        _revertIfHealthFactorBroken(msg.sender);
        s_ausd.mint(msg.sender, _amount);
    }

    function depositCollateralAndMintAUSD(address token, uint256 _collateralAmount, uint256 _aUSDAmount) public onlyAllowedTokens(token) moreThanZero(_collateralAmount) moreThanZero(_aUSDAmount) {
        depositCollateral(token, _collateralAmount);
        mintAUSD(_aUSDAmount);
    }

    function burnAUSD(uint256 _amount) public moreThanZero(_amount) {
        _burnAUSD(msg.sender, msg.sender, _amount);
    }

    function liquidate(address who, uint256 _amount) public moreThanZero(_amount) {}

    function getAccountInformation(address user) public view returns(uint256 totalUSDCollateral, uint256 aUSDDebt) {
        if(user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        aUSDDebt = s_totalDept[user];
        totalUSDCollateral = getTotalCollateralInUSD(user);
    }

    function _burnAUSD(address onBehalfOf, address aUSDFrom, uint256 _amount) private  {
        s_totalDept[onBehalfOf] -= _amount;

        s_ausd.burn(aUSDFrom, _amount);
    }

    function _getHealthFactor(uint256 totalUSDCollateral, uint256 aUSDDebt) private pure returns(uint256 healthFactor) {
        if(aUSDDebt == 0) return type(uint256).max;

        uint256 collateralAdjusted = (totalUSDCollateral * LIQUIDATION_THRESHOLD) / LIQUIDATION_PRECISION;

        healthFactor = (collateralAdjusted * PRECISION) / aUSDDebt;
    }

    function getTotalCollateralInUSD(address user) public view returns(uint256 totalAmount) {
        if(user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        address[] memory tokens = s_tokensAllowed;
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

    function _revertIfHealthFactorBroken(address user) private view {
        (uint256 totalUSDCollateral, uint256 aUSDDebt) = getAccountInformation(user);

        uint256 healthFactor = _getHealthFactor(totalUSDCollateral, aUSDDebt);
        if(healthFactor < MIN_HEALTH_FACTOR) {
            revert AUSDEngine__HealthFactorBroken();
        }
    }

    function setAUSD(AnchorUSD _ausd) public onlyOwner {
        if(address(s_ausd) != address(0)) {
            revert AUSDEngine__AUSDAlreadyDefined();
        }

        s_ausd = _ausd;
    }
}