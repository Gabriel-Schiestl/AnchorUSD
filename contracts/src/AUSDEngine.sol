//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {AnchorUSD} from "./AnchorUSD.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {OracleLib} from "./lib/OracleLib.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title AUSDEngine
 * @author Gabriel Schiestl
 * @notice This contract is the core of the AnchorUSD stablecoin system
 *
 * @dev The system is designed to be minimal and maintain a 1 token == $1 USD peg at all times.
 *
 * Stablecoin Properties:
 * - Exogenously Collateralized: Uses external collateral (WETH, WBTC, etc)
 * - Dollar Pegged: Maintains 1:1 parity with USD
 * - Algorithmically Stable: Stability maintained through on-chain mechanics
 *
 * It is similar to DAI, but without governance, no fees, and backed only by WETH and WBTC.
 *
 * CRITICAL INVARIANT: The system must ALWAYS be "overcollateralized".
 * At no point should the value of all collateral be less than the dollar value of all minted AUSD.
 *
 * @notice This contract manages all the logic for:
 * - Minting and burning AUSD
 * - Depositing and withdrawing collateral
 * - Liquidations of unhealthy positions
 * - Calculating user Health Factors
 *
 * @notice Based on the MakerDAO DSS (DAI Stablecoin System)
 */
contract AUSDEngine is ReentrancyGuard {
    //////// Errors ////////

    /// @notice Error thrown when someone other than the owner tries to execute a restricted function
    error AUSDEngine__NotOwner();

    /// @notice Error thrown when attempting to set the AUSD address more than once
    error AUSDEngine__AUSDAlreadyDefined();

    /// @notice Error thrown when a zero address is provided where not allowed
    error AUSDEngine__NotZeroAddress();

    /// @notice Error thrown when the oracle price is invalid (negative)
    error AUSDEngine__InvalidPrice();

    /// @notice Error thrown when attempting to use a token that is not allowed as collateral
    error AUSDEngine__TokenNotAllowed();

    /// @notice Error thrown when a value must be greater than zero but is not
    error AUSDEngine__MustBeMoreThanZero();

    /// @notice Error thrown when there is insufficient collateral for the operation
    error AUSDEngine__InsufficientCollateral();

    /// @notice Error thrown when an operation would break the user's health factor
    error AUSDEngine__HealthFactorBroken();

    /// @notice Error thrown when attempting to liquidate a user with a healthy health factor
    error AUSDEngine__HealthFactorOk();

    /// @notice Error thrown when liquidation does not improve the user's health factor
    error AUSDEngine__HealthFactorNotImproved();

    /// @notice Error thrown when token and price feed arrays have different sizes
    error AUSDEngine__TokenAddressesAndPriceFeedAddressesAmountsDontMatch();

    /// @notice Error thrown when amount of tokens to burn is higher than the user's debt
    error AUSDEngine__BurnAmountExceedsDebt();

    // Type Declarations
    /// @dev Using SafeERC20 for safe operations with ERC20 tokens
    using SafeERC20 for IERC20;

    /// @dev Using OracleLib for staleness checks of price feeds
    using OracleLib for AggregatorV3Interface;

    //////// State Variables ////////

    /// @dev Reference to the AnchorUSD stablecoin contract
    AnchorUSD private s_ausd;

    /// @dev Additional precision to adjust Chainlink prices (8 decimals) to 18 decimals
    /// Chainlink price feeds return values with 8 decimals, we multiply by 1e10 to get 18 decimals
    uint256 private constant PRICE_ADITIONAL_PRECISION = 1e10;

    /// @dev Standard precision used throughout the system (18 decimals, WEI format)
    uint256 private constant PRECISION = 1e18;

    /// @dev Liquidation threshold: 50 means user needs to be 200% collateralized
    /// That is, for every $100 of AUSD minted, needs to have at least $200 in collateral
    uint256 private constant LIQUIDATION_THRESHOLD = 50;

    /// @dev Precision used for calculations involving the liquidation threshold
    uint256 private constant LIQUIDATION_PRECISION = 100;

    /// @dev Minimum allowed health factor (1e18 = 1.0)
    /// If health factor < MIN_HEALTH_FACTOR, the user can be liquidated
    uint256 private constant MIN_HEALTH_FACTOR = 1e18;

    /// @dev Bonus given to liquidator: 10% means liquidator receives 110% of the covered value
    /// For example: covers $100 of debt, receives $110 in collateral
    uint256 private constant LIQUIDATION_BONUS = 10;

    /// @dev Contract owner, the only one capable of setting the AUSD token address
    address private immutable i_owner;

    /// @dev Nested mapping: user => token => amount of collateral deposited
    /// Tracks how much collateral each user deposited of each token type
    mapping(address user => mapping(address token => uint256 collateral))
        private s_collateralDeposited;

    /// @dev Mapping that tracks the total debt in AUSD of each user
    mapping(address user => uint256 debt) private s_totalDebt;

    /// @dev Mapping that associates each collateral token with its Chainlink price feed
    mapping(address token => address priceFeed) private s_priceFeeds;

    /// @dev Array with all tokens allowed as collateral
    /// Used to iterate over all collateral types when calculating total value
    address[] private s_tokensAllowed;

    //////// Events ////////

    /**
     * @notice Emitted when a user deposits collateral
     * @param user Address of the user who deposited
     * @param token Address of the token deposited as collateral
     * @param amount Amount deposited
     */
    event CollateralDeposited(
        address indexed user,
        address indexed token,
        uint256 indexed amount
    );

    /**
     * @notice Emitted when collateral is redeemed
     * @param user Address of the user who redeemed the collateral
     * @param token Address of the redeemed token
     * @param amount Amount redeemed
     */
    event CollateralRedeemed(
        address indexed user,
        address indexed token,
        uint256 indexed amount
    );

    /**
     * @notice Emitted when AUSD is minted
     * @param user Address of the user who received the AUSD
     * @param amount Amount of AUSD minted
     */
    event AUSDMinted(address indexed user, uint256 indexed amount);

    /**
     * @notice Emitted when AUSD is burned
     * @param user Address of the user whose debt was reduced
     * @param amount Amount of AUSD burned
     */
    event AUSDBurned(address indexed user, uint256 indexed amount);

    /**
     * @notice Emitted when a liquidation occurs
     * @param liquidatedUser Address of the user being liquidated
     * @param liquidator Address of the liquidator
     * @param tokenCollateral Address of the collateral token transferred
     * @param collateralAmount Amount of collateral transferred (including bonus)
     * @param debtCovered Amount of debt covered in AUSD
     */
    event Liquidation(
        address indexed liquidatedUser,
        address indexed liquidator,
        address indexed tokenCollateral,
        uint256 collateralAmount,
        uint256 debtCovered
    );

    //////// Modifiers ////////

    /**
     * @notice Restricts function to contract owner
     * @dev Reverts with AUSDEngine__NotOwner if msg.sender is not the owner
     */
    modifier onlyOwner() {
        if (msg.sender != i_owner) {
            revert AUSDEngine__NotOwner();
        }
        _;
    }

    /**
     * @notice Ensures the token is allowed as collateral
     * @param token Address of the token to be verified
     * @dev Reverts with AUSDEngine__TokenNotAllowed if the token doesn't have a configured price feed
     */
    modifier onlyAllowedTokens(address token) {
        if (s_priceFeeds[token] == address(0)) {
            revert AUSDEngine__TokenNotAllowed();
        }
        _;
    }

    /**
     * @notice Ensures the value is greater than zero
     * @param amount Value to be verified
     * @dev Reverts with AUSDEngine__MustBeMoreThanZero if amount == 0
     */
    modifier moreThanZero(uint256 amount) {
        if (amount == 0) {
            revert AUSDEngine__MustBeMoreThanZero();
        }
        _;
    }

    //////// Functions ////////

    /**
     * @notice Initializes the contract with collateral tokens and their price feeds
     * @param tokenAddresses Array with addresses of ERC20 tokens allowed as collateral
     * @param priceFeedAddresses Array with addresses of corresponding Chainlink price feeds
     *
     * @dev Arrays must have the same size and corresponding indices
     * @dev Sets msg.sender as contract owner
     * @dev Validates that no address is zero
     * @dev Configures the price feeds mapping and populates the allowed tokens array
     *
     * Requirements:
     * - tokenAddresses and priceFeedAddresses must have the same size
     * - No address can be zero
     */
    constructor(
        address[] memory tokenAddresses,
        address[] memory priceFeedAddresses
    ) {
        i_owner = msg.sender;

        if (tokenAddresses.length != priceFeedAddresses.length) {
            revert AUSDEngine__TokenAddressesAndPriceFeedAddressesAmountsDontMatch();
        }

        for (uint256 i = 0; i < tokenAddresses.length; i++) {
            if (
                (tokenAddresses[i] == address(0)) ||
                (priceFeedAddresses[i] == address(0))
            ) {
                revert AUSDEngine__NotZeroAddress();
            }

            s_priceFeeds[tokenAddresses[i]] = priceFeedAddresses[i];
            s_tokensAllowed.push(tokenAddresses[i]);
        }
    }

    //////// Public Functions ////////

    /**
     * @notice Deposits collateral into the system
     * @param token Address of the ERC20 token to be deposited as collateral
     * @param _amount Amount of tokens to deposit (in base units of the token)
     *
     * @dev Updates the s_collateralDeposited mapping with the new balance
     * @dev Emits CollateralDeposited event
     * @dev Uses SafeERC20 for secure transfer
     * @dev Protected against reentrancy with nonReentrant
     *
     * Requirements:
     * - token must be allowed (onlyAllowedTokens)
     * - _amount must be > 0 (moreThanZero)
     * - user must have approved the contract to spend _amount tokens
     *
     * Effects:
     * - Increases s_collateralDeposited[msg.sender][token]
     * - Transfers tokens from user to contract
     *
     * @custom:security Protected against reentrancy
     */
    function depositCollateral(
        address token,
        uint256 _amount
    ) public onlyAllowedTokens(token) moreThanZero(_amount) nonReentrant {
        s_collateralDeposited[msg.sender][token] += _amount;
        emit CollateralDeposited(msg.sender, token, _amount);

        IERC20(token).safeTransferFrom(msg.sender, address(this), _amount);
    }

    /**
     * @notice Deposits collateral and mints AUSD in a single transaction
     * @param token Address of the collateral token
     * @param _collateralAmount Amount of collateral to deposit
     * @param _aUSDAmount Amount of AUSD to mint
     *
     * @dev Convenience function that combines depositCollateral and mintAUSD
     * @dev Execution order: first deposits, then mints (to ensure collateralization)
     *
     * Requirements:
     * - token must be allowed
     * - Both _collateralAmount and _aUSDAmount must be > 0
     * - After minting, health factor must be above MIN_HEALTH_FACTOR
     */
    function depositCollateralAndMintAUSD(
        address token,
        uint256 _collateralAmount,
        uint256 _aUSDAmount
    )
        public
        onlyAllowedTokens(token)
        moreThanZero(_collateralAmount)
        moreThanZero(_aUSDAmount)
    {
        depositCollateral(token, _collateralAmount);
        mintAUSD(_aUSDAmount);
    }

    /**
     * @notice Redeems (withdraws) collateral from the system
     * @param token Address of the collateral token to redeem
     * @param _amount Amount to redeem
     *
     * @dev Calls _redeemCollateral internally
     * @dev Checks health factor after redemption
     * @dev Protected against reentrancy
     *
     * Requirements:
     * - token must be allowed
     * - _amount must be > 0
     * - user must have sufficient collateral deposited
     * - health factor after redemption must be OK
     *
     * @custom:security Checks health factor after operation to prevent under-collateralization
     */
    function redeemCollateral(
        address token,
        uint256 _amount
    ) public onlyAllowedTokens(token) moreThanZero(_amount) nonReentrant {
        _redeemCollateral(msg.sender, token, _amount);
        _revertIfHealthFactorBroken(msg.sender);
    }

    /**
     * @notice Redeems collateral and burns AUSD in a single transaction
     * @param token Address of the collateral token
     * @param _collateralAmount Amount of collateral to redeem
     * @param aUSDToBurn Amount of AUSD to burn
     *
     * @dev Execution order: first burns AUSD, then redeems collateral
     * @dev This ensures the position improves before releasing collateral
     * @dev Protected against reentrancy
     *
     * Requirements:
     * - token must be allowed
     * - Both values must be > 0
     * - user must have sufficient AUSD to burn
     * - user must have sufficient collateral to redeem
     * - final health factor must be OK
     */
    function redeemCollateralForAUSD(
        address token,
        uint256 _collateralAmount,
        uint256 aUSDToBurn
    )
        public
        onlyAllowedTokens(token)
        moreThanZero(_collateralAmount)
        moreThanZero(aUSDToBurn)
        nonReentrant
    {
        _burnAUSD(msg.sender, aUSDToBurn);
        _redeemCollateral(msg.sender, token, _collateralAmount);
        _revertIfHealthFactorBroken(msg.sender);
    }

    /**
     * @notice Mints (creates) new AUSD tokens
     * @param _amount Amount of AUSD to mint
     *
     * @dev Increases user debt and checks health factor
     * @dev Protected against reentrancy
     *
     * Requirements:
     * - _amount must be > 0
     * - user must have sufficient collateral
     * - health factor after minting must be above MIN_HEALTH_FACTOR
     *
     * Effects:
     * - Increases s_totalDebt[msg.sender]
     * - Calls s_ausd.mint() to create tokens
     *
     * @custom:security Critical: checks health factor BEFORE minting to prevent under-collateralization
     */
    function mintAUSD(
        uint256 _amount
    ) public moreThanZero(_amount) nonReentrant {
        s_totalDebt[msg.sender] += _amount;
        emit AUSDMinted(msg.sender, _amount);

        _revertIfHealthFactorBroken(msg.sender);
        s_ausd.mint(msg.sender, _amount);
    }

    /**
     * @notice Burns (destroys) AUSD tokens
     * @param _amount Amount of AUSD to burn
     *
     * @dev Reduces user debt
     *
     * Requirements:
     * - _amount must be > 0
     * - user must have sufficient AUSD
     *
     * Effects:
     * - Decreases s_totalDebt[msg.sender]
     * - Burns AUSD tokens
     */
    function burnAUSD(uint256 _amount) public moreThanZero(_amount) {
        _burnAUSD(msg.sender, _amount);
    }

    /**
     * @notice Liquidates an unhealthy position (health factor < 1)
     * @param user Address of the user to be liquidated
     * @param token Address of the collateral token to receive
     * @param debtToCover Amount of debt (AUSD) to cover
     *
     * @dev Liquidation mechanism is fundamental to maintaining protocol solvency
     * @dev Liquidator pays the debt in AUSD and receives collateral + 10% bonus
     * @dev Protected against reentrancy
     *
     * How it works:
     * 1. Verifies that user has health factor < MIN_HEALTH_FACTOR
     * 2. Calculates amount of collateral equivalent to debtToCover
     * 3. Adds 10% bonus to collateral
     * 4. Transfers collateral from user to liquidator
     * 5. Burns AUSD from liquidator
     * 6. Verifies that user's health factor improved
     * 7. Verifies that liquidator maintains healthy health factor
     *
     * Requirements:
     * - user cannot be address(0)
     * - token must be allowed
     * - debtToCover must be > 0
     * - user must have health factor < MIN_HEALTH_FACTOR
     * - user must have sufficient collateral of the specified token
     * - liquidator must have sufficient AUSD
     * - user's health factor must improve after liquidation
     * - liquidator's health factor must remain healthy
     *
     * @custom:security Critical: multiple checks to prevent improper liquidations
     * @custom:security Checks liquidator's health factor at the end
     * @custom:note Known bug: if protocol is only 100% collateralized (without over-collateralization),
     *                        liquidations may not be possible
     */
    function liquidate(
        address user,
        address token,
        uint256 debtToCover
    ) public onlyAllowedTokens(token) moreThanZero(debtToCover) nonReentrant {
        if (user == address(0)) revert AUSDEngine__NotZeroAddress();

        uint256 startingHealthFactor = _getHealthFactor(user);
        if (startingHealthFactor >= MIN_HEALTH_FACTOR)
            revert AUSDEngine__HealthFactorOk();

        if (s_collateralDeposited[user][token] == 0)
            revert AUSDEngine__InsufficientCollateral();

        uint256 tokenAmount = getTokenAmountFromUSD(token, debtToCover);

        uint256 bonusCollateral = ((tokenAmount * LIQUIDATION_BONUS) /
            LIQUIDATION_PRECISION);

        uint256 totalCollateralToRedeem = tokenAmount + bonusCollateral;

        if (s_collateralDeposited[user][token] < totalCollateralToRedeem) {
            revert AUSDEngine__InsufficientCollateral();
        }
        s_collateralDeposited[user][token] -= totalCollateralToRedeem;

        if (s_totalDebt[user] < debtToCover)
            revert AUSDEngine__BurnAmountExceedsDebt();
        s_totalDebt[user] -= debtToCover;

        emit Liquidation(
            user,
            msg.sender,
            token,
            totalCollateralToRedeem,
            debtToCover
        );

        IERC20(token).safeTransfer(msg.sender, totalCollateralToRedeem);
        s_ausd.burn(msg.sender, debtToCover);

        uint256 endingHealthFactor = _getHealthFactor(user);
        if (endingHealthFactor <= startingHealthFactor)
            revert AUSDEngine__HealthFactorNotImproved();

        _revertIfHealthFactorBroken(msg.sender);
    }

    //////// Private Functions ////////

    /**
     * @notice Gets user account information
     * @param user User address
     * @return totalUSDCollateral Total collateral value in USD
     * @return aUSDDebt Total debt in AUSD
     *
     * @dev Helper function used internally for calculations
     * @dev Validates that user is not address(0)
     */
    function _getAccountInformation(
        address user
    ) private view returns (uint256 totalUSDCollateral, uint256 aUSDDebt) {
        if (user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        aUSDDebt = s_totalDebt[user];
        totalUSDCollateral = getTotalCollateralInUSD(user);
    }

    /**
     * @notice Internal function to redeem collateral
     * @param user Address of the user redeeming collateral
     * @param token Address of the collateral token
     * @param _amount Amount to redeem
     *
     * @dev Emits CollateralRedeemed event
     * @dev Uses SafeERC20 for secure transfer
     *
     * Requirements:
     * - user must have sufficient collateral
     *
     * Effects:
     * - Decreases s_collateralDeposited[user][token]
     * - Transfers tokens to user
     *
     * @custom:security Critical: does not check health factor - caller function's responsibility
     */
    function _redeemCollateral(
        address user,
        address token,
        uint256 _amount
    ) private {
        uint256 collateralDeposited = s_collateralDeposited[user][token];
        if (_amount > collateralDeposited) {
            revert AUSDEngine__InsufficientCollateral();
        }

        s_collateralDeposited[user][token] -= _amount;
        emit CollateralRedeemed(user, token, _amount);

        IERC20(token).safeTransfer(user, _amount);
    }

    /**
     * @notice Internal function to burn AUSD
     * @param user Address whose debt will be reduced and AUSD burned from
     * @param _amount Amount of AUSD to burn
     *
     * @dev Emits AUSDBurned event
     *
     * Requirements:
     * - user must have debt >= _amount
     * - user must have AUSD >= _amount
     *
     * Effects:
     * - Decreases s_totalDebt[user]
     * - Burns AUSD tokens from user
     */
    function _burnAUSD(address user, uint256 _amount) private {
        if (s_totalDebt[user] < _amount)
            revert AUSDEngine__BurnAmountExceedsDebt();

        s_totalDebt[user] -= _amount;

        emit AUSDBurned(user, _amount);

        s_ausd.burn(user, _amount);
    }

    //////// Private & Internal View & Pure Functions ////////

    /**
     * @notice Calculates health factor based on collateral and debt
     * @param totalUSDCollateral Total collateral value in USD
     * @param aUSDDebt Total debt in AUSD
     * @return healthFactor The calculated health factor
     *
     * @dev Formula: healthFactor = (collateralAdjusted * PRECISION) / debt
     * @dev collateralAdjusted = (totalUSDCollateral * LIQUIDATION_THRESHOLD) / LIQUIDATION_PRECISION
     * @dev LIQUIDATION_THRESHOLD = 50 means only 50% of collateral counts for health factor
     * @dev This creates the need for 200% collateralization (1 / 0.5 = 2)
     *
     * @dev Returns type(uint256).max if there is no debt (prevents division by zero)
     *
     * Example:
     * - Collateral: $200 USD
     * - Debt: $100 AUSD
     * - collateralAdjusted = 200 * 50 / 100 = $100
     * - healthFactor = 100 * 1e18 / 100 = 1e18 (exactly at the limit)
     *
     * - To be safe, healthFactor must be >= 1e18 (1.0)
     * - healthFactor < 1e18 allows liquidation
     */
    function _calculateHealthFactor(
        uint256 totalUSDCollateral,
        uint256 aUSDDebt
    ) private pure returns (uint256 healthFactor) {
        if (aUSDDebt == 0) return type(uint256).max;

        uint256 collateralAdjusted = (totalUSDCollateral *
            LIQUIDATION_THRESHOLD) / LIQUIDATION_PRECISION;

        healthFactor = (collateralAdjusted * PRECISION) / aUSDDebt;
    }

    /**
     * @notice Calculates a user's health factor
     * @param user User address
     * @return healthFactor The user's health factor
     *
     * @dev Gets account information and calls _calculateHealthFactor
     */
    function _getHealthFactor(
        address user
    ) private view returns (uint256 healthFactor) {
        (uint256 totalUSDCollateral, uint256 aUSDDebt) = _getAccountInformation(
            user
        );

        healthFactor = _calculateHealthFactor(totalUSDCollateral, aUSDDebt);
    }

    /**
     * @notice Gets a token's price from the Chainlink price feed
     * @param token Token address
     * @return Token price in USD with 8 decimals
     *
     * @dev Uses OracleLib.staleCheckLatestRoundData() to check for staleness
     * @dev Validates that the price is not negative
     *
     * @custom:security Critical: uses staleCheckLatestRoundData to prevent use of stale data
     */
    function _getPrice(address token) private view returns (int256) {
        AggregatorV3Interface priceFeed = AggregatorV3Interface(
            s_priceFeeds[token]
        );

        (, int256 price, , , ) = priceFeed.staleCheckLatestRoundData();

        if (price < 0) {
            revert AUSDEngine__InvalidPrice();
        }

        return price;
    }

    /**
     * @notice Calculates the USD value of a user's collateral for a specific token
     * @param token Collateral token address
     * @param user User address
     * @return Collateral value in USD (18 decimals)
     *
     * @dev Formula: (price * PRICE_ADITIONAL_PRECISION * collateral) / PRECISION
     * @dev PRICE_ADITIONAL_PRECISION converts price from 8 to 18 decimals
     *
     * Example:
     * - 1 ETH deposited
     * - ETH Price = $2000 (returned as 2000_00000000 = 2000 * 1e8)
     * - priceAdjusted = 2000_00000000 * 1e10 = 2000 * 1e18
     * - collateral = 1 * 1e18 (1 ETH in wei)
     * - usdCollateral = (2000 * 1e18 * 1 * 1e18) / 1e18 = 2000 * 1e18
     * - Result: $2000 USD (in 18 decimal format)
     */
    function _getCollateralInUSD(
        address token,
        address user
    ) private view returns (uint256) {
        int256 price = _getPrice(token);

        uint256 priceAdjusted = uint256(price) * PRICE_ADITIONAL_PRECISION;

        uint256 collateral = s_collateralDeposited[user][token];

        uint256 usdCollateral = (priceAdjusted * collateral) / PRECISION;

        return usdCollateral;
    }

    /**
     * @notice Reverts the transaction if the user's health factor is broken
     * @param user Address of the user to verify
     *
     * @dev Critical security function called after operations that may affect health factor
     * @dev Reverts with AUSDEngine__HealthFactorBroken if healthFactor < MIN_HEALTH_FACTOR
     *
     * @custom:security Critical: ensures all operations maintain protocol solvency
     */
    function _revertIfHealthFactorBroken(address user) private view {
        uint256 healthFactor = _getHealthFactor(user);
        if (healthFactor < MIN_HEALTH_FACTOR) {
            revert AUSDEngine__HealthFactorBroken();
        }
    }

    //////// External & Public View & Pure Functions ////////

    /**
     * @notice Returns the liquidation threshold
     * @return LIQUIDATION_THRESHOLD (50 = 50%, requires 200% collateralization)
     */
    function getLiquidationThreshold() external pure returns (uint256) {
        return LIQUIDATION_THRESHOLD;
    }

    /**
     * @notice Returns the precision used for liquidation calculations
     * @return LIQUIDATION_PRECISION (100)
     */
    function getLiquidationPrecision() external pure returns (uint256) {
        return LIQUIDATION_PRECISION;
    }

    /**
     * @notice Returns the minimum allowed health factor
     * @return MIN_HEALTH_FACTOR (1e18 = 1.0)
     */
    function getMinHealthFactor() external pure returns (uint256) {
        return MIN_HEALTH_FACTOR;
    }

    /**
     * @notice Returns the liquidation bonus
     * @return LIQUIDATION_BONUS (10 = 10% bonus for liquidators)
     */
    function getLiquidationBonus() external pure returns (uint256) {
        return LIQUIDATION_BONUS;
    }

    /**
     * @notice Returns the system's standard precision
     * @return PRECISION (1e18 = WEI format)
     */
    function getPrecision() external pure returns (uint256) {
        return PRECISION;
    }

    /**
     * @notice Returns the additional precision used to adjust prices
     * @return PRICE_ADITIONAL_PRECISION (1e10)
     */
    function getPriceAdditionalPrecision() external pure returns (uint256) {
        return PRICE_ADITIONAL_PRECISION;
    }

    /**
     * @notice Calculates the total value of a user's collateral in USD
     * @param user User address
     * @return totalAmount Total value in USD (18 decimals)
     *
     * @dev Iterates over all allowed tokens and sums their values
     * @dev Validates that user is not address(0)
     */
    function getTotalCollateralInUSD(
        address user
    ) public view returns (uint256 totalAmount) {
        if (user == address(0)) {
            revert AUSDEngine__NotZeroAddress();
        }

        address[] memory tokens = s_tokensAllowed;
        for (uint256 i = 0; i < tokens.length; i++) {
            totalAmount += _getCollateralInUSD(tokens[i], user);
        }
    }

    /**
     * @notice Returns the caller's health factor
     * @return msg.sender's health factor
     */
    function getHealthFactor() public view returns (uint256) {
        return _getHealthFactor(msg.sender);
    }

    /**
     * @notice Returns a user's collateral balance for a specific token
     * @param user User address
     * @param token Token address
     * @return Amount deposited by the user
     */
    function getCollateralBalanceOfUser(
        address user,
        address token
    ) external view returns (uint256) {
        return s_collateralDeposited[user][token];
    }

    /**
     * @notice Returns a user's total debt
     * @param user User address
     * @return Total debt in AUSD
     */
    function getUserDebt(address user) external view returns (uint256) {
        return s_totalDebt[user];
    }

    /**
     * @notice Returns a specific user's health factor
     * @param user User address
     * @return User's health factor
     */
    function getUserHealthFactor(address user) external view returns (uint256) {
        return _getHealthFactor(user);
    }

    /**
     * @notice Returns complete account information for a user
     * @param user User address
     * @return totalUSDCollateral Total collateral value in USD
     * @return aUSDDebt Total debt in AUSD
     */
    function getUserAccountInformation(
        address user
    ) external view returns (uint256 totalUSDCollateral, uint256 aUSDDebt) {
        return _getAccountInformation(user);
    }

    /**
     * @notice Converts a USD value to token amount
     * @param token Token address
     * @param _amount Value in USD (18 decimals)
     * @return Equivalent token amount
     *
     * @dev Used primarily to calculate how much collateral to give to liquidator
     * @dev Formula: (_amount * PRECISION) / (price * PRICE_ADITIONAL_PRECISION)
     *
     * Requirements:
     * - token must be allowed
     *
     * Example:
     * - Cover $100 USD of debt
     * - ETH Price = $2000
     * - tokenAmount = (100 * 1e18 * 1e18) / (2000 * 1e8 * 1e10)
     * - tokenAmount = 0.05 * 1e18 (0.05 ETH)
     */
    function getTokenAmountFromUSD(
        address token,
        uint256 _amount
    ) public view onlyAllowedTokens(token) returns (uint256) {
        int256 price = _getPrice(token);

        return ((_amount * PRECISION) /
            (uint256(price) * PRICE_ADITIONAL_PRECISION));
    }

    /**
     * @notice Returns the address of a token's price feed
     * @param token Token address
     * @return Chainlink price feed address
     */
    function getTokenPriceFeed(address token) external view returns (address) {
        return s_priceFeeds[token];
    }

    /**
     * @notice Returns array with all tokens allowed as collateral
     * @return Array of allowed token addresses
     */
    function getAllowedTokens() external view returns (address[] memory) {
        return s_tokensAllowed;
    }

    /**
     * @notice Returns the contract owner's address
     * @return Owner's address
     */
    function getOwner() external view returns (address) {
        return i_owner;
    }

    /**
     * @notice Returns the price of a collateral token in USD
     * @param token Token address
     * @return Price in USD (18 decimals)
     *
     * @dev Returns price already adjusted to 18 decimals
     */
    function getCollateralTokenPrice(
        address token
    ) external view returns (uint256) {
        int256 price = _getPrice(token);
        return uint256(price) * PRICE_ADITIONAL_PRECISION;
    }

    //////// Setters (Owner Only) ////////

    /**
     * @notice Sets the AnchorUSD contract address
     * @param _ausd AnchorUSD contract instance
     *
     * @dev Can only be called once by the owner
     * @dev Necessary to establish bidirectional connection between contracts
     *
     * Requirements:
     * - Only owner can call
     * - s_ausd must not have been set yet
     *
     * @custom:security Critical: ensures AUSD can only be set once
     */
    function setAUSD(AnchorUSD _ausd) public onlyOwner {
        if (address(s_ausd) != address(0)) {
            revert AUSDEngine__AUSDAlreadyDefined();
        }

        s_ausd = _ausd;
    }
}
