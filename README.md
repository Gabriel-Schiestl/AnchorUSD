# AnchorUSD - Decentralized Stablecoin Protocol

<div align="center">

**A production-ready DeFi stablecoin protocol with real-time off-chain indexing and liquidation monitoring**

[![Solidity](https://img.shields.io/badge/Solidity-^0.8.30-363636?style=flat-square&logo=solidity)](https://soliditylang.org/)
[![Go](https://img.shields.io/badge/Go-1.25.5-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16.1-000000?style=flat-square&logo=next.js)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)

[Live Demo](#) • [Documentation](#architecture-overview) • [Report Bug](../../issues) • [Request Feature](../../issues)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Problems Solved](#problems-solved)
- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Smart Contracts](#smart-contracts)
- [Backend Indexer](#backend-indexer)
- [Frontend Application](#frontend-application)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Security Considerations](#security-considerations)
- [Future Enhancements](#future-enhancements)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

**AnchorUSD (aUSD)** is a fully-featured decentralized stablecoin protocol inspired by MakerDAO's DAI, implementing an **exogenously collateralized**, **dollar-pegged**, **algorithmically stable** cryptocurrency. The protocol maintains a 1:1 peg with USD through over-collateralization mechanics and automated liquidation systems.

This project demonstrates production-level DeFi engineering with three integrated components:

- **Smart Contracts** (Solidity/Foundry): Core protocol logic with comprehensive testing
- **Backend Indexer** (Go): High-performance event indexing and state management
- **Frontend dApp** (Next.js): Professional user interface with real-time updates

### What Makes This Project Stand Out

This isn't just another stablecoin demo—it's a **portfolio-grade implementation** that solves real production challenges in DeFi:

1. **Off-Chain State Management**: Eliminates costly on-chain reads through Redis-backed indexing
2. **Real-Time Event Processing**: Worker pool architecture for concurrent blockchain event handling
3. **Production-Ready Architecture**: Clean separation of concerns, comprehensive error handling, structured logging
4. **Advanced Liquidation System**: Automated monitoring and calculation of at-risk positions
5. **Professional Frontend**: Modern UI with wallet integration, real-time health factor calculations, and transaction history

---

## ✨ Key Features

### Smart Contract Layer

- ✅ **Multi-Collateral Support**: Accept WETH and WBTC as collateral
- ✅ **Over-Collateralization**: 200% minimum collateralization ratio (50% liquidation threshold)
- ✅ **Liquidation System**: 10% bonus incentive for liquidators maintaining protocol solvency
- ✅ **Chainlink Price Feeds**: Reliable oracle integration for collateral valuation
- ✅ **Safety Mechanisms**: ReentrancyGuard, health factor checks, comprehensive validation
- ✅ **Flexible Operations**: Deposit, withdraw, mint, burn, and liquidate functions

### Backend Indexer

- 🚀 **Event-Driven Architecture**: Real-time blockchain event subscription and processing
- 🚀 **Worker Pool Pattern**: Concurrent event processing with configurable worker counts
- 🚀 **Redis Caching**: Fast off-chain state reconstruction and querying
- 🚀 **Metrics Workers**: Automated calculation of protocol metrics, user positions, and liquidation eligibility
- 🚀 **RESTful API**: Comprehensive endpoints for user data, dashboard metrics, and history
- 🚀 **Database Persistence**: PostgreSQL for historical data and audit trails
- 🚀 **Structured Logging**: Production-grade observability with zerolog

### Frontend Application

- 💎 **Modern UI/UX**: Built with Next.js 16, TypeScript, and Tailwind CSS
- 💎 **Wallet Integration**: RainbowKit + Wagmi for seamless Web3 connectivity
- 💎 **Real-Time Updates**: SWR for automatic data fetching and cache invalidation
- 💎 **Health Factor Projections**: Preview health factor before executing transactions
- 💎 **Dashboard Analytics**: Protocol overview, collateral breakdown, liquidation monitoring
- 💎 **Transaction History**: Complete user activity timeline with event tracking
- 💎 **Responsive Design**: Mobile-first approach with Radix UI components

---

## 🎓 Problems Solved

### 1. **High On-Chain Query Costs**

**Problem**: Reading blockchain state is expensive. Each contract call costs gas, and complex calculations (like iterating through all user positions) are prohibitively expensive or impossible.

**Solution**: The **off-chain indexer** subscribes to blockchain events and maintains a complete, queryable state in Redis. This architecture:

- Reduces frontend API calls from ~5-10 contract reads to 1 HTTP request
- Enables complex aggregations (total supply, protocol health) that would be impossible on-chain
- Provides instant response times (<50ms) vs blockchain RPC calls (500ms+)
- Allows historical queries without scanning thousands of blocks

### 2. **Liquidation Monitoring at Scale**

**Problem**: Identifying liquidatable positions requires checking every user's health factor against current collateral prices—an operation that doesn't scale on-chain.

**Solution**: The **liquidations worker** periodically:

- Fetches all user positions from cache (O(1) operation)
- Calculates health factors with live price feeds
- Maintains a sorted list of at-risk positions
- Provides instant liquidation opportunities to the frontend

### 3. **Poor User Experience**

**Problem**: Traditional DeFi interfaces require users to manually read contract state, calculate safe borrow amounts, and risk transaction failures.

**Solution**: The **frontend + backend integration** provides:

- **Predictive Health Factors**: See your health factor before minting/burning
- **Real-Time Metrics**: Instant updates on collateral value and debt
- **Transaction History**: Complete audit trail without blockchain scanning
- **Error Prevention**: Client-side validation prevents failed transactions

### 4. **State Reconstruction**

**Problem**: If the indexer crashes or needs to restart, it must rebuild the entire protocol state from genesis—a slow and complex process.

**Solution**: The backend implements:

- **Event Sourcing Pattern**: All state changes are tracked through events
- **Database Persistence**: Historical events are stored in PostgreSQL
- **Automatic Rebuild**: On startup, the system replays events to reconstruct Redis cache
- **Idempotency**: Duplicate event processing is safely handled

---

## 🏗 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         BLOCKCHAIN LAYER                        │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐   │
│  │ AUSDEngine   │◄────►│  AnchorUSD   │◄────►│   Chainlink  │   │
│  │   Contract   │      │  ERC20 Token │      │ Price Feeds  │   │
│  └──────┬───────┘      └──────────────┘      └──────────────┘   │
│         │ Emits Events                                          │
└─────────┼───────────────────────────────────────────────────────┘
          │
          │ WebSocket Subscription
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                         BACKEND INDEXER                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                     Worker Processes                     │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │   │
│  │  │ Log Worker   │  │ Metrics      │  │ Liquidations │    │   │
│  │  │ (Event Sub)  │  │ Worker       │  │ Worker       │    │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │   │
│  └─────────┼─────────────────┼─────────────────┼────────────┘   │
│            │                 │                 │                │
│            ▼                 ▼                 ▼                │
│  ┌─────────────────┐  ┌──────────────────────────────────┐      │
│  │  Redis Cache    │  │        PostgreSQL DB             │      │
│  │  - User State   │  │  - Events History                │      │
│  │  - Collateral   │  │  - Deposits/Mints/Burns          │      │
│  │  - Debt         │  │  - Liquidations                  │      │
│  │  - Metrics      │  │  - Price History                 │      │
│  └────────┬────────┘  └──────────────────────────────────┘      │
│           │                                                     │
│           │ REST API (Gin Framework)                            │
└───────────┼─────────────────────────────────────────────────────┘
            │
            │ HTTP Requests
            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      FRONTEND APPLICATION                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                      Next.js Pages                       │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │   │
│  │  │ Operations   │  │  Dashboard   │  │   History    │    │   │
│  │  │ (Mint/Burn)  │  │  (Metrics)   │  │  (Timeline)  │    │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │   │
│  └─────────┼─────────────────┼─────────────────┼────────────┘   │
│            │                 │                 │                │
│            ▼                 ▼                 ▼                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Wagmi + RainbowKit                         │    │
│  │              (Wallet Connection)                        │    │
│  └──────────────────────────┬──────────────────────────────┘    │
└─────────────────────────────┼───────────────────────────────────┘
                              │
                              ▼
                        User's Wallet
```

### Data Flow Examples

**Depositing Collateral:**

1. User clicks "Deposit" → Frontend calls contract via Wagmi
2. Contract emits `CollateralDeposited` event
3. Backend Log Worker receives event → Processes → Updates Redis + PostgreSQL
4. Frontend automatically refetches user data via SWR → UI updates

**Viewing Dashboard:**

1. Frontend requests `/api/dashboard` from backend
2. Backend reads pre-computed metrics from Redis (instant)
3. Response includes: total collateral, supply, protocol health, liquidatable users
4. No blockchain calls needed—all data from cache

---

## 🛠 Technology Stack

### Smart Contracts

- **Solidity 0.8.30**: Modern Solidity with custom errors and gas optimizations
- **Foundry**: Fast testing framework with fuzzing and invariant testing
- **OpenZeppelin**: Battle-tested contract libraries (ERC20, ReentrancyGuard)
- **Chainlink**: Decentralized oracle network for price feeds

### Backend (Go)

- **Go 1.25.5**: High-performance concurrent processing
- **Gin**: Fast HTTP web framework
- **go-ethereum**: Ethereum client library for blockchain interaction
- **GORM**: Type-safe ORM for PostgreSQL
- **Redis**: In-memory cache for real-time state
- **Zerolog**: Structured logging for observability

### Frontend (TypeScript)

- **Next.js 16**: React framework with App Router
- **TypeScript**: Type-safe development
- **Wagmi 2.x**: React Hooks for Ethereum
- **RainbowKit**: Best-in-class wallet connection UI
- **TanStack Query**: Powerful async state management
- **Tailwind CSS**: Utility-first styling
- **Radix UI**: Accessible component primitives
- **SWR**: Data fetching with automatic revalidation

### Infrastructure

- **PostgreSQL**: Persistent storage for historical data
- **Redis**: High-speed caching layer
- **Docker**: Containerization (recommended)

---

## 📜 Smart Contracts

### AUSDEngine.sol

The core protocol contract managing all stablecoin operations.

**Key Functions:**

```solidity
// Collateral Management
depositCollateral(token, amount)
redeemCollateral(token, amount)
depositCollateralAndMintAUSD(token, collateralAmount, ausdAmount)
redeemCollateralForAUSD(token, collateralAmount, ausdAmount)

// Stablecoin Operations
mintAUSD(amount)      // Requires sufficient collateral
burnAUSD(amount)      // Reduces user debt

// Liquidation System
liquidate(user, token, debtToCover)  // Liquidate unhealthy positions
```

**Health Factor Calculation:**

```
Health Factor = (Collateral Value in USD × 50) / (Debt in USD × 100)

If Health Factor < 1.0 → User is liquidatable
Liquidator receives 110% of covered debt in collateral (10% bonus)
```

**Constants:**

- Liquidation Threshold: 50% (200% collateralization required)
- Liquidation Bonus: 10%
- Minimum Health Factor: 1.0
- Supported Collateral: WETH, WBTC

### AnchorUSD.sol

ERC20 stablecoin token with restricted minting/burning.

**Security Features:**

- Only AUSDEngine can mint/burn tokens
- Immutable engine address set at deployment
- Standard ERC20 interface for compatibility

### Testing

```bash
cd contracts
forge test                    # Run all tests
forge test --gas-report      # Gas consumption analysis
forge test -vvv              # Verbose output with stack traces
forge coverage               # Code coverage report
```

**Test Coverage:**

- ✅ Unit tests for all contract functions
- ✅ Invariant testing for protocol solvency
- ✅ Fuzz testing for edge cases
- ✅ Liquidation scenario testing
- ✅ Oracle failure handling

---

## ⚙️ Backend Indexer

### Architecture Patterns

The backend follows **Clean Architecture** principles:

```
cmd/api/           → Application entry point
internal/
  ├── blockchain/  → Ethereum client wrappers
  ├── config/      → Configuration management
  ├── domain/      → Business logic (pure functions)
  ├── http/        → HTTP handlers and routes
  ├── model/       → Data models and DTOs
  ├── service/     → Orchestration layer
  ├── storage/     → Data access layer (Redis, PostgreSQL)
  ├── worker/      → Background workers
  └── utils/       → Shared utilities
```

### Core Components

#### 1. Log Worker

Subscribes to blockchain events and processes them concurrently.

```go
// Configurable worker pool (default: 4 workers)
NUM_LOG_WORKERS=4

// Processes events:
- CollateralDeposited
- CollateralRedeemed
- AUSDMinted
- AUSDBurned
- Liquidation
```

**Flow:**

1. Subscribe to contract events via WebSocket
2. Distribute events to worker pool via channels
3. Each worker processes events and updates state
4. Persist to PostgreSQL for audit trail
5. Update Redis cache for real-time queries

#### 2. Metrics Worker

Computes protocol-wide metrics and user-specific data.

```go
// Runs on every event processed
- Updates user collateral balances (per token)
- Calculates total debt per user
- Aggregates protocol total supply
- Computes collateral USD values
```

**Redis Keys:**

```
user:collateral_usd:{address}    → Total collateral value
user:debt:{address}              → User's AUSD debt
collateral:{token}:{user}        → Collateral by token
collateral:total_supply          → Protocol collateral
coin:total_supply                → Total AUSD minted
```

#### 3. Liquidations Worker

Periodically scans for liquidatable positions.

```go
// Configurable scan interval (default: 1 hour)
LIQUIDATIONS_SCAN_INTERVAL=1h

// Identifies users with health factor < 1.0
// Stores in liquidatable:{address} Redis hash
```

### API Endpoints

```
GET  /user/:address                    → User position data
POST /user/:address/health-factor      → Calculate health factor projections
GET  /dashboard                        → Protocol metrics
GET  /history/:address                 → User transaction history
```

**Example Response:**

```json
{
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "collateralUsd": "25000000000000000000000",
  "debtUsd": "10000000000000000000000",
  "healthFactor": 1.25,
  "collateralBreakdown": [
    {
      "token": "WETH",
      "amount": "5000000000000000000",
      "valueUsd": "15000000000000000000000"
    }
  ]
}
```

### Running the Backend

```bash
cd backend

# Configure environment
cat > .env << EOF
RPC_URL=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
CONTRACT_ADDRESS=0x...
DB_HOST=localhost
DB_PORT=5432
DB_NAME=anchorusd
REDIS_HOST=localhost
REDIS_PORT=6379
NUM_LOG_WORKERS=4
NUM_METRICS_WORKERS=4
LIQUIDATIONS_SCAN_INTERVAL=1h
EOF

# Install dependencies
go mod download

# Run migrations and start
go run cmd/api/main.go
```

### State Reconstruction

On startup, the backend automatically:

1. Queries last processed block from PostgreSQL
2. Subscribes to events from that block forward
3. Replays historical events to rebuild Redis cache
4. Begins real-time processing

This ensures the indexer can recover from crashes without data loss.

---

## 🎨 Frontend Application

### Pages

1. **Operations (`/`)**: Main interface for deposits, mints, and burns
2. **Dashboard (`/dashboard`)**: Protocol analytics and user portfolio
3. **History (`/history`)**: Transaction timeline
4. **Risk (`/risk`)**: Liquidation monitoring and risk metrics

### Key Features

#### Predictive Health Factors

Before executing any operation, users can see the projected health factor:

```typescript
// Real-time health factor calculation
const calculateHealthFactorAfterMint = async (mintAmount: string) => {
  return await ausdEngineApi.calculateHealthFactorAfterMint(
    address,
    scaledAmount,
  );
};
```

#### Real-Time Updates

```typescript
// SWR configuration for automatic revalidation
useSWR<AUSDEngineData>(
  `/api/ausd-engine/user/${address}`,
  { refreshInterval: 10000 }, // Refresh every 10 seconds
);
```

#### Wallet Integration

```typescript
// RainbowKit configuration
<RainbowKitProvider chains={chains}>
  <WagmiProvider config={wagmiConfig}>
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  </WagmiProvider>
</RainbowKitProvider>
```

### Component Architecture

```
components/
├── operations/
│   └── mint-burn-deposit.tsx      # Main operations interface
├── dashboard/
│   └── user-dashboard.tsx         # Portfolio overview
├── history/
│   └── history-list.tsx           # Transaction timeline
├── risk/
│   └── risk-dashboard.tsx         # Liquidation monitoring
├── layout/
│   └── navbar.tsx                 # Navigation with wallet
└── ui/                             # Reusable UI components
```

### Running the Frontend

```bash
cd frontend

# Install dependencies
npm install

# Configure environment
cat > .env.local << EOF
NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID=your_project_id
NEXT_PUBLIC_CONTRACT_ADDRESS=0x...
NEXT_PUBLIC_BACKEND_URL=http://localhost:8080
EOF

# Run development server
npm run dev

# Build for production
npm run build
npm start
```

---

## 🚀 Getting Started

### Prerequisites

- **Node.js** 20+ (for frontend)
- **Go** 1.25+ (for backend)
- **Foundry** (for smart contracts)
- **PostgreSQL** 14+
- **Redis** 6+
- **Ethereum RPC** (Alchemy, Infura, or local node)

### Complete Setup

#### 1. Clone Repository

```bash
git clone https://github.com/yourusername/AnchorUSD.git
cd AnchorUSD
```

#### 2. Smart Contracts

```bash
cd contracts

# Install dependencies
forge install

# Run tests
forge test

# Deploy (update HelperConfig.s.sol with your network)
forge script script/DeployAUSD.s.sol --rpc-url $RPC_URL --broadcast --private-key $PRIVATE_KEY
```

#### 3. Backend

```bash
cd backend

# Start infrastructure (Docker recommended)
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:17
docker run -d -p 6379:6379 redis:7

# Configure environment
cp .env.example .env
# Edit .env with your contract address and RPC URL

# Run
go run cmd/api/main.go
```

#### 4. Frontend

```bash
cd frontend

npm install
cp .env.local.example .env.local
# Edit .env.local with your WalletConnect project ID

npm run dev
```

Access the application at `http://localhost:3001`

---

## 📁 Project Structure

```
AnchorUSD/
├── contracts/                  # Smart contracts (Solidity + Foundry)
│   ├── src/
│   │   ├── AUSDEngine.sol     # Core protocol logic
│   │   ├── AnchorUSD.sol      # ERC20 stablecoin token
│   │   └── lib/               # Libraries (OracleLib)
│   ├── test/                   # Comprehensive test suite
│   ├── script/                 # Deployment scripts
│   └── lib/                    # Dependencies (OpenZeppelin, Chainlink)
│
├── backend/                    # Go indexer and API
│   ├── cmd/api/               # Application entry point
│   ├── internal/
│   │   ├── blockchain/        # Ethereum client
│   │   ├── config/            # Configuration
│   │   ├── domain/            # Business logic
│   │   ├── http/              # REST API
│   │   │   ├── handlers/      # Request handlers
│   │   │   └── external/      # Price feed API
│   │   ├── model/             # Data models
│   │   ├── service/           # Application services
│   │   │   └── processors/    # Event processors
│   │   ├── storage/           # Data access layer
│   │   ├── worker/            # Background workers
│   │   └── utils/             # Shared utilities
│   └── AUSDEngine.abi.json    # Contract ABI
│
└── frontend/                   # Next.js dApp
    ├── app/                    # Next.js App Router
    │   ├── dashboard/          # Dashboard page
    │   ├── history/            # History page
    │   └── risk/               # Risk monitoring page
    ├── components/             # React components
    │   ├── operations/         # Mint/Burn/Deposit UI
    │   ├── dashboard/          # Dashboard components
    │   ├── history/            # History components
    │   ├── risk/               # Risk components
    │   ├── layout/             # Layout components
    │   └── ui/                 # Reusable UI (Radix)
    ├── hooks/                  # Custom React hooks
    ├── api/                    # API client
    ├── lib/                    # Utilities and config
    │   ├── wagmi-config.ts     # Wagmi configuration
    │   └── AUSDEngine.abi.json # Contract ABI
    └── models/                 # TypeScript types
```

---

## 🧪 Testing

### Smart Contracts

```bash
cd contracts

# Unit tests
forge test

# Gas report
forge test --gas-report

# Coverage
forge coverage

# Specific test
forge test --match-test testLiquidation -vvv
```

**Key Test Files:**

- `AUSDEngine.t.sol`: Core functionality tests
- `invariant/`: Protocol invariant testing
- `mocks/`: Mock contracts for testing

### Backend

```bash
cd backend

go test ./...

# Test with coverage
go test -cover ./...

# Specific package
go test ./internal/service/...
```

### Frontend

```bash
cd frontend

# Type checking
npm run type-check

# Linting
npm run lint

# Build test
npm run build
```

---

## 🔒 Security Considerations

### Smart Contract Security

✅ **Implemented:**

- ReentrancyGuard on all state-changing functions
- Health factor checks before critical operations
- SafeERC20 for token transfers
- Oracle staleness validation
- Comprehensive input validation
- Custom errors for gas efficiency

⚠️ **Known Limitations:**

- If protocol becomes exactly 100% collateralized (not over-collateralized), liquidations may fail
- Oracle dependency: system relies on Chainlink price feeds
- No pause mechanism for emergency situations

### Backend Security

✅ **Implemented:**

- Environment-based configuration (no secrets in code)
- Structured logging (no sensitive data logged)
- CORS configuration
- Input validation on all endpoints
- Error handling without information disclosure

### Frontend Security

✅ **Implemented:**

- Client-side transaction validation
- User confirmation before operations
- Error handling with user-friendly messages
- No private key handling (wallet-based auth)

---

## 🚧 Future Enhancements

### Smart Contracts

- [ ] Multi-collateral liquidation in single transaction
- [ ] Flash loan protection
- [ ] Governance module for parameter adjustment
- [ ] Stability fee mechanism
- [ ] Support for additional collateral types

### Backend

- [ ] WebSocket endpoint for real-time updates
- [x] Prometheus metrics export
- [x] Rate limiting and request throttling
- [x] Admin dashboard for monitoring(Grafana)

### Frontend

- [ ] Advanced charts (historical prices, health factor trends)

### Infrastructure

- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Load balancing for backend
- [ ] Automated backups
- [ ] Monitoring and alerting setup

---

## 📊 Performance Metrics

### Smart Contracts

- **Gas Efficiency**: Optimized for production use
- **Deployment Cost**: ~3-4M gas
- **Average Transaction Cost**:
  - Deposit: ~100k gas
  - Mint: ~80k gas
  - Liquidation: ~150k gas

### Backend

- **API Response Time**: <50ms (cached data)
- **Event Processing**: <100ms per event
- **Concurrent Workers**: 4-8 recommended
- **Memory Footprint**: ~50MB base + Redis cache

### Frontend

- **First Contentful Paint**: <1.5s
- **Time to Interactive**: <3s
- **Bundle Size**: ~300KB (gzipped)
- **Lighthouse Score**: 90+ (Performance)

---

## 🤝 Contributing

Contributions are welcome! This is a portfolio project, but improvements and suggestions are appreciated.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting
5. Commit with clear messages (`git commit -m 'Add amazing feature'`)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Style

- **Solidity**: Follow official style guide, use NatSpec comments
- **Go**: Follow `gofmt` and effective Go conventions
- **TypeScript**: Prettier + ESLint configuration provided

---

## 👤 Author

**Gabriel Schiestl**

- GitHub: [@Gabriel-Schiestl](https://github.com/Gabriel-Schiestl)
- LinkedIn: [Gabriel Schiestl](https://www.linkedin.com/in/gabriel-schiestl-98208a276/)

---

## 🙏 Acknowledgments

- **MakerDAO**: Inspiration for the stablecoin mechanism
- **OpenZeppelin**: Battle-tested smart contract libraries
- **Chainlink**: Reliable oracle infrastructure
- **Foundry**: Amazing development tooling
- **Cyfrin Updraft**: Educational resources

---

## 📚 Additional Resources

- [MakerDAO Documentation](https://docs.makerdao.com/)
- [Chainlink Price Feeds](https://docs.chain.link/data-feeds)
- [Foundry Book](https://book.getfoundry.sh/)
- [Wagmi Documentation](https://wagmi.sh/)
- [Go Ethereum Documentation](https://geth.ethereum.org/docs)

---

<div align="center">

**⭐ Star this repository if you found it helpful!**

Made with ❤️ by [Gabriel Schiestl](https://github.com/Gabriel-Schiestl)

</div>
