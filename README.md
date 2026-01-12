# AnchorUSD (aUSD) - Algorithmic Stablecoin Protocol

<div align="center">

![Solidity](https://img.shields.io/badge/Solidity-0.8.30-363636?style=for-the-badge&logo=solidity)
![Foundry](https://img.shields.io/badge/Foundry-Framework-orange?style=for-the-badge)
![Chainlink](https://img.shields.io/badge/Chainlink-Oracles-375BD2?style=for-the-badge&logo=chainlink)
![OpenZeppelin](https://img.shields.io/badge/OpenZeppelin-Security-4E5EE4?style=for-the-badge)

**A decentralized crypto-collateralized stablecoin protocol with automated liquidation mechanism**

[Architecture](#-architecture) • [Key Features](#-key-features) • [Security](#-security) • [Installation](#-installation) • [Deployment](#-deployment)

</div>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Architecture](#-architecture)
- [Key Features](#-key-features)
- [System Components](#-system-components)
- [Security](#-security)
- [Technologies](#-technologies)
- [Installation](#-installation)
- [Deployment](#-deployment)
- [Testing](#-testing)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## 🎯 Overview

**AnchorUSD (aUSD)** is a DeFi protocol implementing a decentralized algorithmic stablecoin collateralized by crypto assets (WETH and WBTC). The protocol utilizes Chainlink oracles to ensure accurate pricing and implements a robust liquidation system to maintain overcollateralization and system stability.

### 🎓 Highlights for Web3 Recruiters

This project demonstrates proficiency in:

- **DeFi Protocol Design**: Complete implementation of a stablecoin protocol with CDP (Collateralized Debt Position) mechanics
- **Oracle Integration**: Advanced use of Chainlink Price Feeds with stale data protection
- **Security Best Practices**: Reentrancy guards, checks-effects-interactions pattern, custom errors for gas optimization
- **Smart Contract Patterns**: Factory pattern, upgradeable architecture, modular design
- **Mathematical Precision**: Collateralization, health factor and liquidation calculations with decimal precision
- **Foundry Expertise**: Use of modern framework for development, testing and deployment

---

## 🏗️ Architecture

The protocol follows a modular and secure design with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                        User Interface                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                      AUSDEngine.sol                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  • Deposit/Redeem Collateral (WETH, WBTC)            │   │
│  │  • Mint/Burn aUSD                                    │   │
│  │  • Liquidation Mechanism                             │   │
│  │  • Health Factor Calculations                        │   │
│  └──────────────────────────────────────────────────────┘   │
└───────┬─────────────────────────┬───────────────────────────┘
        │                         │
        ▼                         ▼
┌─────────────────┐      ┌──────────────────────┐
│  AnchorUSD.sol  │      │   OracleLib.sol      │
│   (ERC20)       │      │  (Price Feeds)       │
│                 │      │                      │
│  • Mint/Burn    │      │  • Stale Check       │
│  • Access       │      │  • Price Feed        │
│    Control      │      │    Validation        │
└─────────────────┘      └──────────┬───────────┘
                                    │
                         ┌──────────▼───────────┐
                         │  Chainlink Oracles   │
                         │  ETH/USD, BTC/USD    │
                         └──────────────────────┘
```

### Implemented Design Patterns

1. **Access Control Pattern**: The aUSD token can only be minted/burned by the engine
2. **Oracle Pattern**: Separate library for price feed validation
3. **Checks-Effects-Interactions**: Reentrancy prevention in all critical functions
4. **Factory Pattern**: Automated deployment with network configurations via HelperConfig

---

## ✨ Key Features

### 💰 Collateralization System

- **Multiple Collateral**: Support for WETH and WBTC as collateral
- **Liquidation Ratio**: 50% (200% overcollateralization ratio)
- **Health Factor**: Continuous monitoring system for position health
- **Incentivized Liquidation**: 10% bonus for liquidators

### 🔄 Core Functionalities

```solidity
// Deposit collateral and mint aUSD in a single transaction
depositCollateralAndMintAUSD(token, collateralAmount, aUSDAmount)

// Burn aUSD and redeem collateral
redeemCollateralForAUSD(token, collateralAmount, aUSDToBurn)

// Liquidate unhealthy positions
liquidate(user, token, debtToCover)
```

### 📊 Metrics and Monitoring

- **Health Factor Calculation**: `(collateralValue * 50) / 100 / totalDebt`
- **Minimum Health Factor**: 1e18 (1.0 at 18 decimal scale)
- **Price Precision**: 1e18 for precise calculations
- **Liquidation Bonus**: 10% of liquidated value

---

## 🧩 System Components

### 1. AnchorUSD.sol (Token Contract)

ERC20 implementation of the stablecoin with restricted access control:

```solidity
// Key characteristics
- Mint/Burn only by the engine
- No custom transfer functions
- Immutable engine address for security
- Custom errors for gas efficiency
```

**Relevant Technical Points**:

- Use of `immutable` for gas optimization
- Restricted access pattern with custom errors
- Clean implementation without unnecessary functionality

### 2. AUSDEngine.sol (Core Logic)

Main protocol engine with all critical operations:

**State Management**:

```solidity
mapping(address user => mapping(address token => uint256)) s_collateralDeposited
mapping(address user => uint256 debt) s_totalDebt
mapping(address token => address priceFeed) s_priceFeeds
```

**Main Functions**:

| Function              | Description                           | Restrictions                            |
| --------------------- | ------------------------------------- | --------------------------------------- |
| `depositCollateral()` | Deposits collateral into the protocol | NonReentrant, Allowed token             |
| `mintAUSD()`          | Issues new aUSD tokens                | Validates health factor                 |
| `redeemCollateral()`  | Withdraws collateral                  | Validates health factor post-withdrawal |
| `liquidate()`         | Liquidates unhealthy position         | Only if health factor < 1.0             |

**Implemented Security**:

- ✅ ReentrancyGuard on all transfer functions
- ✅ SafeERC20 for safe transfers
- ✅ Health factor validation before and after operations
- ✅ Custom errors for gas savings
- ✅ Modifiers for input validation

### 3. OracleLib.sol (Oracle Integration)

Custom library for secure Chainlink integration:

```solidity
// Fresh data validation (2 hour timeout)
function staleCheckLatestRoundData() returns (uint80, int256, uint256, uint256, uint80)

// Security features
- updatedAt timestamp verification
- answeredInRound vs roundId validation
- Configurable timeout (2 hours)
- Reverts with informative custom errors
```

**Why is this important?**:

- Prevents use of stale prices in volatile market conditions
- Protects against oracle failures or manipulations
- Ensures on-chain data integrity

---

## 🔒 Security

### Implemented Security Measures

#### 1. **Reentrancy Protection**

```solidity
- OpenZeppelin ReentrancyGuard on all critical functions
- Checks-Effects-Interactions pattern
- State updates before external calls
```

#### 2. **Oracle Validation**

```solidity
- OracleLib with stale price detection
- 2-hour timeout for price data
- roundId and answeredInRound validation
- Negative price rejection
```

#### 3. **Access Control**

```solidity
- onlyOwner for administrative functions
- onlyEngine on token contract
- Immutable addresses for critical components
```

#### 4. **Input Validation**

```solidity
- moreThanZero modifier
- onlyAllowedTokens modifier
- Zero address checks
- Array length validations
```

#### 5. **Economic Security**

```solidity
- 200% Overcollateralization
- 50% Liquidation threshold
- Liquidation bonus to incentivize liquidators
- Continuous health factor monitoring
```

### Future Security Considerations

For production, consider:

- [ ] Professional security audit (Consensys Diligence, Trail of Bits, OpenZeppelin)
- [ ] Bug bounty program
- [ ] Time-locks on administrative functions
- [ ] Multi-sig wallet for owner
- [ ] Circuit breakers for emergency situations
- [ ] Stress testing under extreme market conditions

---

## 🛠️ Technologies

### Blockchain & Smart Contracts

- **Solidity 0.8.30**: Programming language for smart contracts
- **Foundry**: Modern development and testing framework
  - Forge: Compilation and testing
  - Cast: Blockchain interaction
  - Anvil: Local node for development

### Libraries & Integrations

- **OpenZeppelin Contracts**:
  - ERC20 implementation
  - ReentrancyGuard
  - SafeERC20
- **Chainlink**:
  - AggregatorV3Interface for price feeds
  - MockV3Aggregator for testing

### Oracles

- **Chainlink Price Feeds**:
  - ETH/USD
  - BTC/USD
  - Decentralized price data

---

## 📦 Installation

### Prerequisites

```bash
# Foundry (Forge, Cast, Anvil)
curl -L https://foundry.paradigm.xyz | bash
foundryup

# Git
git --version
```

### Project Setup

```bash
# Clone the repository
git clone https://github.com/your-username/AnchorUSD.git
cd AnchorUSD

# Install dependencies
forge install

# Compile contracts
forge build

# Verify installation
forge test
```

### Folder Structure

```
AnchorUSD/
├── src/
│   ├── AnchorUSD.sol          # ERC20 Token
│   ├── AUSDEngine.sol         # Core logic
│   └── lib/
│       └── OracleLib.sol      # Oracle integration
├── script/
│   ├── DeployAUSD.s.sol       # Deploy script
│   └── HelperConfig.s.sol     # Network configs
├── test/
│   ├── unit/                  # Unit tests (in development)
│   ├── fuzz/                  # Fuzz testing (in development)
│   └── invariant/             # Invariant tests (in development)
├── lib/                       # Dependencies (git submodules)
└── foundry.toml              # Foundry configuration
```

---

## 🚀 Deployment

### Local Deployment (Anvil)

```bash
# 1. Start local node
anvil

# 2. Deploy in new terminal window
forge script script/DeployAUSD.s.sol:DeployAUSD --rpc-url http://localhost:8545 --broadcast
```

### Testnet Deployment (Sepolia)

```bash
# 1. Configure .env
echo "PRIVATE_KEY=your_private_key" > .env
echo "SEPOLIA_RPC_URL=your_alchemy_or_infura_url" >> .env

# 2. Source .env
source .env

# 3. Deploy
forge script script/DeployAUSD.s.sol:DeployAUSD \
    --rpc-url $SEPOLIA_RPC_URL \
    --broadcast \
    --verify \
    -vvvv
```

### Contract Interaction

```bash
# Deposit collateral
cast send $AUSD_ENGINE "depositCollateral(address,uint256)" \
    $WETH_ADDRESS \
    1000000000000000000 \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY

# Mint aUSD
cast send $AUSD_ENGINE "mintAUSD(uint256)" \
    500000000000000000000 \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY

# Check health factor
cast call $AUSD_ENGINE "getHealthFactor()" \
    --rpc-url $RPC_URL
```

---

## 🧪 Testing

### Test Structure (In Development)

The project will include a complete test suite:

#### 1. **Unit Tests** ✅ (Planned)

```bash
forge test --match-path test/unit/*.t.sol
```

Planned coverage:

- ✅ Collateral deposit and withdrawal tests
- ✅ aUSD mint and burn tests
- ✅ Health factor validation
- ✅ Liquidation mechanics
- ✅ Oracle validation
- ✅ Access control

#### 2. **Fuzz Testing** 🔄 (Planned)

```bash
forge test --match-path test/fuzz/*.t.sol
```

Planned scenarios:

- Tests with random collateral values
- Oracle price variations
- Multiple sequential operations
- Numerical edge cases

#### 3. **Invariant Testing** 🔄 (Planned)

```bash
forge test --match-path test/invariant/*.t.sol
```

Protocol invariants:

- aUSD total supply ≤ Total collateral in USD
- Health factor always calculated correctly
- Sum of individual collaterals = total collateral
- Liquidations always improve health factor

### Coverage Report

```bash
# Generate coverage report
forge coverage

# Detailed coverage
forge coverage --report lcov
```

### Gas Profiling

```bash
# Gas report for optimizations
forge test --gas-report
```

---

## 📈 Roadmap

### Phase 1: Foundation ✅ (Complete)

- [x] Core protocol implementation
- [x] Chainlink oracle integration
- [x] Deploy scripts for multiple networks
- [x] Initial documentation

### Phase 2: Testing & Security 🔄 (In Progress)

- [ ] Complete unit test suite
- [ ] Fuzz testing implementation
- [ ] Invariant testing
- [ ] Internal security audit

### Phase 3: Optimization 📋 (Planned)

- [ ] Gas optimization
- [ ] Upgrade to ERC-4626 standard (Tokenized Vaults)
- [ ] Web graphical interface
- [ ] Expanded technical documentation

### Phase 4: Production 📋 (Planned)

- [ ] External security audit
- [ ] Mainnet deployment
- [ ] Bug bounty program
- [ ] Decentralized governance

---

## 💡 Advanced Concepts Demonstrated

### 1. Mathematical Finance

- Health factor calculation: `(collateral * LTV) / debt`
- Overflow/underflow prevention with Solidity 0.8+
- Decimal precision with 18 decimals

### 2. DeFi Mechanics

- Collateralized Debt Positions (CDP)
- Liquidation incentives
- Oracle price feeds
- Overcollateralization

### 3. Gas Optimization

- Custom errors vs require strings (~99% gas saving)
- Immutable variables
- Efficient storage patterns
- View functions for queries

### 4. Security Patterns

- Checks-Effects-Interactions
- Pull over Push payments
- Oracle manipulation resistance
- Access control layers

---

## 🤝 Contributing

Contributions are welcome! To contribute:

1. Fork the project
2. Create a branch for your feature (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

This project is under the MIT license. See the `LICENSE` file for more details.

---

## 📞 Contact

**Developer**: Gabriel Schiestl

- GitHub: [@Gabriel-Schiestl](https://github.com/Gabriel-Schiestl)
- LinkedIn: [Gabriel Schiestl](https://www.linkedin.com/in/gabriel-schiestl-98208a276/)

---

## 🎓 Additional Resources

### Technical Documentation

- [Chainlink Price Feeds](https://docs.chain.link/data-feeds)
- [OpenZeppelin Contracts](https://docs.openzeppelin.com/contracts)
- [Foundry Book](https://book.getfoundry.sh/)

### DeFi Concepts

- [MakerDAO: CDP Model](https://makerdao.com/en/)
- [Collateralized Stablecoins](https://ethereum.org/en/stablecoins/)

---

<div align="center">

**⭐ If this project was useful, consider giving it a star!**

Made with ❤️ for the Web3 community

</div>
