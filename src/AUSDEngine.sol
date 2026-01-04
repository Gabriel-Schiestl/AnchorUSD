//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {AnchorUSD} from "./AnchorUSD.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

contract AUSDEngine {
    //Errors
    error AUSDEngine__NotOwner();
    error AUSDEngine__AUSDAlreadyDefined();
    error AUSDEngine__NotZeroAddress();
    error AUSDEngine__InvalidPrice();
    error AUSDEngine__TokenNotAllowed();
    error AUSDEngine__MustBeMoreThanZero();
    error AUSDEngine__InsufficientCollateral();
    error AUSDEngine__HealthFactorBroken();
    error AUSDEngine__HealthFactorOk();
    error AUSDEngine__HealthFactorNotImproved();
    error AUSDEngine__TokenAddressesAndPriceFeedAddressesAmountsDontMatch();

    //Types
    using SafeERC20 for IERC20;

    //State Variables
    AnchorUSD private s_ausd;

    uint256 private constant PRICE_ADITIONAL_PRECISION = 1e10;
    uint256 private constant PRECISION = 1e18;
    uint256 private constant LIQUIDATION_THRESHOLD = 50;
    uint256 private constant LIQUIDATION_PRECISION = 100;
    uint256 private constant MIN_HEALTH_FACTOR = 1e18;
    uint256 private constant LIQUIDATION_BONUS = 10;

    address private immutable i_owner;
    mapping(address user => mapping(address token => uint256 collateral)) private s_collateralDeposited;
    mapping(address user => uint256 debt) private s_totalDebt;
    mapping(address token => address priceFeed) private s_priceFeeds;
    address[] private s_tokensAllowed;

    //Events
    event CollateralDeposited(address indexed user, address indexed token, uint256 indexed amount);
    event CollateralRedeemed(address indexed redeemedFrom, address indexed redeemedTo, address token, uint256 amount);
    event AUSDMinted(address indexed user, uint256 indexed amount);

    //Modifiers
    modifier onlyOwner() {
        if (msg.sender != i_owner) {
            revert AUSDEngine__NotOwner();
        }
        _;
    }

    modifier onlyAllowedTokens(address token) {
        if (s_priceFeeds[token] == address(0)) {
            revert AUSDEngine__TokenNotAllowed();
        }
        _;
    }

    modifier moreThanZero(uint256 amount) {
        if (amount == 0) {
            revert AUSDEngine__MustBeMoreThanZero();
        }
        _;
    }

    //Functions
    constructor(address[] memory tokenAddresses, address[] memory priceFeedAddresses) {
        i_owner = msg.sender;

        if (tokenAddresses.length != priceFeedAddresses.length) {
            revert AUSDEngine__TokenAddressesAndPriceFeedAddressesAmountsDontMatch();
        }

        for (uint256 i = 0; i < tokenAddresses.length; i++) {
            s_priceFeeds[tokenAddresses[i]] = priceFeedAddresses[i];
            s_tokensAllowed.push(tokenAddresses[i]);
        }
    }

    //Public Functions
    function depositCollateral(address token, uint256 _amount) public onlyAllowedTokens(token) moreThanZero(_amount) {
        s_collateralDeposited[msg.sender][token] += _amount;
        emit CollateralDeposited(msg.sender, token, _amount);

        IERC20(token).safeTransferFrom(msg.sender, address(this), _amount);
    }

    function depositCollateralAndMintAUSD(address token, uint256 _collateralAmount, uint256 _aUSDAmount)
        public
        onlyAllowedTokens(token)
        moreThanZero(_collateralAmount)
        moreThanZero(_aUSDAmount)
    {
        depositCollateral(token, _collateralAmount);
        mintAUSD(_aUSDAmount);
    }

    function redeemCollateral(address token, uint256 _amount) public onlyAllowedTokens(token) moreThanZero(_amount) {
        _redeemCollateral(msg.sender, msg.sender, token, _amount);
        _revertIfHealthFactorBroken(msg.sender);
    }

    function redeemCollateralForAUSD(address token, uint256 _collateralAmount, uint256 aUSDToBurn)
        public
        onlyAllowedTokens(token)
        moreThanZero(_collateralAmount)
        moreThanZero(aUSDToBurn)
    {
        _burnAUSD(msg.sender, msg.sender, aUSDToBurn);
        redeemCollateral(token, _collateralAmount);
        _revertIfHealthFactorBroken(msg.sender);
    }

    function mintAUSD(uint256 _amount) public moreThanZero(_amount) {
        s_totalDebt[msg.sender] += _amount;
        emit AUSDMinted(msg.sender, _amount);

        _revertIfHealthFactorBroken(msg.sender);
        s_ausd.mint(msg.sender, _amount);
    }

    function burnAUSD(uint256 _amount) public moreThanZero(_amount) {
        _burnAUSD(msg.sender, msg.sender, _amount);
    }

    function liquidate(address user, address token, uint256 debtToCover)
        public
        onlyAllowedTokens(token)
        moreThanZero(debtToCover)
    {
        if (user == address(0)) revert AUSDEngine__NotZeroAddress();

        uint256 startingHealthFactor = _getHealthFactor(user);
        if (startingHealthFactor >= MIN_HEALTH_FACTOR) revert AUSDEngine__HealthFactorOk();

        if (s_collateralDeposited[user][token] == 0) revert AUSDEngine__InsufficientCollateral();

        uint256 tokenAmount = getTokenAmountFromUSD(token, debtToCover);

        uint256 bonusCollateral = ((tokenAmount * LIQUIDATION_BONUS) / LIQUIDATION_PRECISION);

        _redeemCollateral(user, msg.sender, token, tokenAmount + bonusCollateral);
        _burnAUSD(user, msg.sender, tokenAmount);

        uint256 endingHealthFactor = _getHealthFactor(user);
        if (endingHealthFactor <= startingHealthFactor) revert AUSDEngine__HealthFactorNotImproved();

        _revertIfHealthFactorBroken(msg.sender);
    }

    //Private Functions
    function _getAccountInformation(address user) public view returns (uint256 totalUSDCollateral, uint256 aUSDDebt) {
        if (user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        aUSDDebt = s_totalDebt[user];
        totalUSDCollateral = getTotalCollateralInUSD(user);
    }

    function _redeemCollateral(address from, address to, address token, uint256 _amount) private {
        uint256 collateralDeposited = s_collateralDeposited[from][token];
        if (_amount > collateralDeposited) {
            revert AUSDEngine__InsufficientCollateral();
        }

        s_collateralDeposited[from][token] -= _amount;
        emit CollateralRedeemed(from, to, token, _amount);

        IERC20(token).safeTransfer(to, _amount);
    }

    function _burnAUSD(address onBehalfOf, address aUSDFrom, uint256 _amount) private {
        s_totalDebt[onBehalfOf] -= _amount;

        s_ausd.burn(aUSDFrom, _amount);
    }

    function _calculateHealthFactor(uint256 totalUSDCollateral, uint256 aUSDDebt)
        private
        pure
        returns (uint256 healthFactor)
    {
        if (aUSDDebt == 0) return type(uint256).max;

        uint256 collateralAdjusted = (totalUSDCollateral * LIQUIDATION_THRESHOLD) / LIQUIDATION_PRECISION;

        healthFactor = (collateralAdjusted * PRECISION) / aUSDDebt;
    }

    function _getHealthFactor(address user) private view returns (uint256 healthFactor) {
        (uint256 totalUSDCollateral, uint256 aUSDDebt) = _getAccountInformation(user);

        healthFactor = _calculateHealthFactor(totalUSDCollateral, aUSDDebt);
    }

    function _getPrice(address token) private view returns (int256) {
        AggregatorV3Interface priceFeed = AggregatorV3Interface(s_priceFeeds[token]);

        (, int256 price,,,) = priceFeed.latestRoundData();

        if (price < 0) {
            revert AUSDEngine__InvalidPrice();
        }

        return price;
    }

    function _getCollateralInUSD(address token, address user) private view returns (uint256) {
        int256 price = _getPrice(token);

        uint256 priceAdjusted = uint256(price) * PRICE_ADITIONAL_PRECISION;

        uint256 collateral = s_collateralDeposited[user][token];

        uint256 usdCollateral = (priceAdjusted * collateral) / PRECISION;

        return usdCollateral;
    }

    function _revertIfHealthFactorBroken(address user) private view {
        uint256 healthFactor = _getHealthFactor(user);
        if (healthFactor < MIN_HEALTH_FACTOR) {
            revert AUSDEngine__HealthFactorBroken();
        }
    }

    //Public & External View/Pure Functions
    function getLiquidationThreshold() external pure returns (uint256) {
        return LIQUIDATION_THRESHOLD;
    }

    function getLiquidationPrecision() external pure returns (uint256) {
        return LIQUIDATION_PRECISION;
    }

    function getMinHealthFactor() external pure returns (uint256) {
        return MIN_HEALTH_FACTOR;
    }

    function getLiquidationBonus() external pure returns (uint256) {
        return LIQUIDATION_BONUS;
    }

    function getPrecision() external pure returns (uint256) {
        return PRECISION;
    }

    function getPriceAdditionalPrecision() external pure returns (uint256) {
        return PRICE_ADITIONAL_PRECISION;
    }

    // Getters for User Data
    function getTotalCollateralInUSD(address user) public view returns (uint256 totalAmount) {
        if (user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        address[] memory tokens = s_tokensAllowed;
        for (uint256 i = 0; i < tokens.length; i++) {
            totalAmount += _getCollateralInUSD(tokens[i], user);
        }
    }

    function getHealthFactor() public view returns (uint256) {
        return _getHealthFactor(msg.sender);
    }

    function getCollateralBalanceOfUser(address user, address token) external view returns (uint256) {
        return s_collateralDeposited[user][token];
    }

    function getUserDebt(address user) external view returns (uint256) {
        return s_totalDebt[user];
    }

    function getUserHealthFactor(address user) external view returns (uint256) {
        return _getHealthFactor(user);
    }

    function getUserAccountInformation(address user)
        external
        view
        returns (uint256 totalUSDCollateral, uint256 aUSDDebt)
    {
        return _getAccountInformation(user);
    }

    // Getters for System Data
    function getTokenAmountFromUSD(address token, uint256 _amount)
        public
        view
        onlyAllowedTokens(token)
        returns (uint256)
    {
        int256 price = _getPrice(token);

        return ((_amount * PRECISION) / (uint256(price) * PRICE_ADITIONAL_PRECISION));
    }

    function getTokenPriceFeed(address token) external view returns (address) {
        return s_priceFeeds[token];
    }

    function getAllowedTokens() external view returns (address[] memory) {
        return s_tokensAllowed;
    }

    function getOwner() external view returns (address) {
        return i_owner;
    }

    function getCollateralTokenPrice(address token) external view returns (uint256) {
        int256 price = _getPrice(token);
        return uint256(price) * PRICE_ADITIONAL_PRECISION;
    }

    //Setters
    function setAUSD(AnchorUSD _ausd) public onlyOwner {
        if (address(s_ausd) != address(0)) {
            revert AUSDEngine__AUSDAlreadyDefined();
        }

        s_ausd = _ausd;
    }
}
