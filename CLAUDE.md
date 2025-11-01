# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**NOFX** is a multi-AI automated cryptocurrency futures trading platform that supports multiple exchanges (Binance, Hyperliquid, Aster DEX) and AI models (DeepSeek, Qwen, custom APIs). The system features a Go backend with AI-powered trading logic and a React/TypeScript frontend for monitoring and configuration.

**Key Architecture**: Hybrid fullstack with database-driven configuration (SQLite), multi-trader management, AI decision-making with self-learning feedback loops, and real-time web dashboard.

## Build & Development Commands

### Backend (Go)

```bash
# Install dependencies
go mod download

# Build the binary
go build -o nofx

# Run the backend server (port 8080)
./nofx

# Run tests
go test ./...

# Run specific package tests
go test ./trader
go test ./decision
```

### Frontend (React + TypeScript)

```bash
# Navigate to frontend directory
cd web

# Install dependencies
npm install

# Development server (port 3000, proxies to backend :8080)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Docker Deployment

```bash
# Start all services (backend + frontend + nginx)
docker compose up -d --build

# View logs
docker compose logs -f

# Stop services
docker compose down

# Convenience script (if available)
./start.sh start --build
./start.sh logs
./start.sh stop
```

## High-Level Architecture

### Backend Architecture (Go)

**Core Design Pattern**: Layered architecture with interface-based abstraction for multi-exchange support.

```
┌─────────────────────────────────┐
│   HTTP API (Gin)                │  api/server.go - RESTful endpoints
│   Authentication (JWT + 2FA)    │  auth/auth.go
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│   Business Logic Layer          │
│   - TraderManager               │  manager/trader_manager.go - multi-trader lifecycle
│   - AutoTrader                  │  trader/auto_trader.go - main trading controller
│   - Decision Engine             │  decision/engine.go - AI prompt building + parsing
│   - Logger + Analytics          │  logger/decision_logger.go - performance tracking
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│   Data Layer                    │
│   - SQLite (config.db)          │  config/database.go - CRUD operations
│   - JSON decision logs          │  decision_logs/{trader_id}/
└──────────────┬──────────────────┘
               │
┌──────────────▼──────────────────┐
│   External Integrations         │
│   - Exchange APIs               │  trader/{binance,hyperliquid,aster}_trader.go
│   - AI APIs (MCP)               │  mcp/client.go - DeepSeek/Qwen integration
│   - Market Data                 │  market/data.go - technical indicators
│   - Coin Pool Screening         │  pool/coin_pool.go
└─────────────────────────────────┘
```

**Key Abstraction**: `Trader` interface (trader/interface.go) defines methods like `OpenLong()`, `CloseLong()`, `SetLeverage()`. Each exchange (Binance, Hyperliquid, Aster) implements this interface independently, allowing the core `AutoTrader` controller to be exchange-agnostic.

### Database Schema (SQLite)

```
Users (id, email, password_hash, otp_secret, otp_verified)
  ↓ (FK: user_id)
AI_Models (id, user_id, name, provider, api_key, enabled)
Exchanges (id, user_id, name, type, api_key, secret_key, enabled)
  ↓ (FK: ai_model_id, exchange_id)
Traders (id, user_id, name, ai_model_id, exchange_id, initial_balance, custom_prompt, is_running)
System_Config (key, value)
```

**Important**: As of v3.0.0, trader configuration moved from static `config.json` to SQLite database. Use web interface or API endpoints to manage traders.

### Frontend Architecture (React + TypeScript)

**Tech Stack**: React 18 + TypeScript + Vite + Tailwind CSS + SWR (data fetching) + Zustand (state management)

**Key Components**:
- `AITradersPage.tsx` - Main trader management interface
- `CompetitionPage.tsx` - Multi-AI leaderboard with real-time comparison
- `EquityChart.tsx` - Account equity curve visualization (Recharts)
- `ComparisonChart.tsx` - Multi-trader ROI comparison
- `LoginPage.tsx` / `RegisterPage.tsx` - Authentication UI with OTP support

**Data Flow**: API client (lib/api.ts) → SWR hooks → React components → Real-time UI updates (5-15 second intervals)

### Trading Cycle Flow (Every 3-5 Minutes)

1. **Historical Performance Analysis** - Analyze last 20 cycles, calculate win rate, profit factor, Sharpe ratio
2. **Account Status Check** - Fetch balance, positions, margin usage, daily P&L
3. **Existing Position Analysis** - Get latest market data + technical indicators (RSI, MACD, EMA) for open positions
4. **New Opportunity Evaluation** - Screen candidate coins (AI500 + OI Top or default list), filter low liquidity
5. **AI Decision** - Build system + user prompts with historical feedback, call AI API (DeepSeek/Qwen), parse JSON response with Chain of Thought (CoT)
6. **Trade Execution** - Close existing positions (priority 1), open new positions (priority 2), apply risk checks
7. **Logging & Feedback** - Save complete decision log (CoT + market data + execution results), update performance metrics

**Self-Learning Mechanism**: Historical performance data (win rate, best/worst coins, recent trades) is automatically included in AI prompts, enabling strategy adaptation.

## Important Technical Details

### Exchange-Specific Considerations

**Binance**:
- Subaccounts restricted to ≤5x leverage (main accounts can use up to 20x altcoins, 50x BTC/ETH)
- Auto-fetches LOT_SIZE precision from exchange info
- Handles position side (LONG/SHORT) and margin mode (cross/isolated)

**Hyperliquid**:
- Requires Ethereum private key (remove `0x` prefix) + wallet address
- Decentralized perpetuals with testnet support
- Different API structure than Binance

**Aster DEX**:
- Binance-compatible API (easy migration)
- Requires 3 credentials: main wallet (user), API wallet (signer), API wallet private key
- API Wallet security system (separate from main wallet)

### Risk Management Rules

- **Per-coin position limit**: Altcoins ≤ 1.5x account equity, BTC/ETH ≤ 10x account equity
- **Leverage limits**: Configurable via web interface (default 5x safe for subaccounts)
- **Margin usage**: Total ≤ 90% (AI autonomously decides usage rate)
- **Risk-reward ratio**: Mandatory ≥ 1:2 (stop-loss:take-profit)
- **Position stacking prevention**: No duplicate opening of same coin + direction

### Decision Logging Format

All decisions saved to `decision_logs/{trader_id}/YYYYMMDD_HHMMSS.json`:
- Complete AI Chain of Thought (reasoning process)
- Input prompt (system rules + user data with historical feedback)
- Structured decision JSON (actions, coins, quantities, leverage, stop-loss/take-profit)
- Account state snapshot (balance, positions, margin usage)
- Execution results (success/failure, actual prices, order IDs)

### Configuration Management

**v3.0.0+**: Web-based configuration through SQLite database. Do NOT edit `config.json` for trader setup.

**Still in config.json**:
- `leverage.btc_eth_leverage` / `leverage.altcoin_leverage`
- `use_default_coins` / `default_coins` list
- `api_server_port` (default 8080)
- `max_daily_loss` / `max_drawdown` risk limits
- `jwt_secret` for authentication

**Through Web Interface / API**:
- AI model configurations (`GET/PUT /api/models`)
- Exchange credentials (`GET/PUT /api/exchanges`)
- Trader creation/deletion/control (`POST/DELETE /api/traders`, `POST /api/traders/:id/start|stop`)
- Custom prompts (`PUT /api/traders/:id/prompt`)

## API Endpoints Reference

### Authentication (Public)
- `POST /api/register` - User registration
- `POST /api/login` - Login with JWT
- `POST /api/verify-otp` - OTP verification
- `POST /api/complete-registration` - Complete registration flow

### Configuration (Protected - JWT Required)
- `GET /api/models` - Get AI model configs
- `PUT /api/models` - Update AI model configs
- `GET /api/exchanges` - Get exchange configs
- `PUT /api/exchanges` - Update exchange configs
- `GET /api/supported-models` - List available AI providers
- `GET /api/supported-exchanges` - List available exchanges

### Trader Management (Protected)
- `GET /api/traders` - List all traders
- `POST /api/traders` - Create new trader (requires configured AI model + exchange)
- `DELETE /api/traders/:id` - Delete trader
- `POST /api/traders/:id/start` - Start trader
- `POST /api/traders/:id/stop` - Stop trader
- `PUT /api/traders/:id/prompt` - Update custom prompt

### Trading Data (Protected)
- `GET /api/status?trader_id=xxx` - System status (running state, cycle count)
- `GET /api/account?trader_id=xxx` - Account info (balance, equity, margin)
- `GET /api/positions?trader_id=xxx` - Open positions
- `GET /api/equity-history?trader_id=xxx` - Historical equity for charts
- `GET /api/decisions/latest?trader_id=xxx` - Last 5 decisions
- `GET /api/statistics?trader_id=xxx` - Performance statistics
- `GET /api/performance?trader_id=xxx` - AI learning analysis
- `GET /api/competition` - Multi-trader competition data

### System
- `GET /health` - Health check endpoint

## Testing & Development Workflow

### Adding a New Exchange

1. Create `trader/{exchange_name}_trader.go` implementing the `Trader` interface
2. Add exchange-specific authentication fields to `Exchanges` table schema in `config/database.go`
3. Update `GetExchanges()` and `UpdateExchanges()` in `config/database.go`
4. Add creation logic in `manager/trader_manager.go` `CreateTrader()` switch statement
5. Update `api/server.go` `handleGetExchanges()` and `handleUpdateExchanges()` to include new fields
6. Update frontend `ExchangeIcons.tsx` and exchange configuration UI

### Modifying AI Prompts

**System Prompt** (decision/engine.go): Fixed trading rules, constraints, leverage limits, risk management guidelines

**User Prompt** (decision/engine.go): Dynamic market data, account state, position history, historical performance feedback

To modify prompts, edit `GetFullDecisionWithCustomPrompt()` in `decision/engine.go`. Custom prompts can also be set per-trader via web interface.

### Testing Trading Logic

Use small capital (100-500 USDT recommended) and monitor:
- Backend logs for decision cycles and execution results
- `decision_logs/{trader_id}/` JSON files for complete AI reasoning
- Web dashboard for real-time account/position updates
- Database `config.db` for persistent trader state

### Performance Metrics

Performance tracked in `logger/decision_logger.go`:
- **Win rate**: Profitable trades / total trades
- **Profit/Loss ratio**: Average profit / average loss
- **Sharpe ratio**: Risk-adjusted returns (uses standard deviation)
- **Per-coin statistics**: Win rate and average P/L per coin
- **Recent trade history**: Last 5 trades with entry/exit prices and actual USDT P/L (considers leverage)

**Critical**: v2.0.2+ calculates actual USDT P/L using `Position Value × Price Change % × Leverage` (not just percentages).

## Security Considerations

- JWT tokens for API authentication (secret in `config.json`)
- 2FA/OTP support for user accounts (stored in Users table)
- API keys/secrets stored in SQLite database (consider encryption for production)
- CORS middleware enabled for cross-origin requests (api/server.go)
- Private keys for Hyperliquid/Aster should NEVER be committed to git
- API wallet system for Aster provides additional security layer

## Common Gotchas

1. **Subaccount Leverage**: Binance subaccounts are restricted to ≤5x leverage. Setting higher values will cause trades to fail.

2. **Position Key Format**: System uses `symbol_side` format (e.g., `BTCUSDT_long`) to track positions. This prevents conflicts when holding both long and short on same coin.

3. **Precision Handling**: Exchange info must be fetched to get correct LOT_SIZE precision. System auto-handles this, but network issues can cause precision errors.

4. **Decision Cycle Timing**: Default 3-5 minutes. Shorter intervals may trigger API rate limits; longer intervals may miss opportunities.

5. **Database vs JSON**: v3.0.0+ uses database for configuration. Old `config.json` trader arrays are ignored. Only use web interface or API for trader management.

6. **Historical Feedback**: Performance analysis looks at last 20 cycles. Insufficient historical data will result in limited feedback to AI.

7. **Coin Pool Defaults**: If `use_default_coins: true` or no coin pool API provided, system uses hardcoded list (BTC, ETH, SOL, BNB, XRP, DOGE, ADA, HYPE).

8. **Frontend Proxy**: Vite dev server (port 3000) proxies API requests to backend (port 8080). Check `web/vite.config.ts` for proxy configuration.

## Useful File Paths

**Backend Core**:
- `main.go` - Application entry point, initializes TraderManager
- `trader/auto_trader.go` - Main trading loop (runCycle method)
- `trader/interface.go` - Trader interface definition
- `decision/engine.go` - AI decision logic and prompt building
- `manager/trader_manager.go` - Multi-trader lifecycle management
- `config/database.go` - SQLite schema and CRUD operations
- `api/server.go` - Gin HTTP server and route handlers

**Frontend Core**:
- `web/src/App.tsx` - Main application component
- `web/src/lib/api.ts` - API client wrapper
- `web/src/types.ts` - TypeScript type definitions
- `web/src/components/AITradersPage.tsx` - Trader management UI
- `web/src/components/CompetitionPage.tsx` - Multi-AI leaderboard

**Configuration**:
- `config.json` - System-level configuration (leverage, coin pool, API port)
- `config.db` - SQLite database (created automatically, stores traders/models/exchanges)
- `.env` - Environment variables (if used)

**Documentation**:
- `README.md` - Comprehensive project documentation
- `DOCKER_DEPLOY.md` / `DOCKER_DEPLOY.en.md` - Docker deployment guides
- `INTEGRATION_BOUNTY_*.md` - Exchange integration guides

## Running Tests Workflow

```bash
# Backend tests
go test ./...

# Run with verbose output
go test -v ./trader
go test -v ./decision

# Run specific test
go test -run TestAutoTrader ./trader

# Frontend (if tests exist)
cd web
npm test
```

## Additional Notes

- This is an experimental trading system. Risk management is critical.
- AI decisions do NOT guarantee profitability. Markets are unpredictable.
- System requires stable internet connection for exchange APIs and AI APIs.
- DeepSeek recommended for beginners (cheaper, faster, no VPN needed).
- Logs grow large over time; consider rotation or cleanup scripts.
- SWR caching on frontend reduces redundant API calls (15-second refresh intervals).
- Multi-trader competition mode requires separate exchange accounts/API keys.
