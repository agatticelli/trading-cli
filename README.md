# trading-cli

Production-ready CLI for cryptocurrency futures trading with natural language support, risk management, and multi-account orchestration. Built on a modular architecture with 5 independent Go modules.

## Features

### Core Trading
- **Multi-Account Support**: Manage unlimited trading accounts simultaneously
- **Risk-Based Position Sizing**: Automatic position calculation based on risk percentage
- **Advanced Order Types**: Limit, market, stop loss, take profit, trailing stops
- **Multi-Level TP/SL**: Partial position closing with multiple take profit levels
- **Break Even**: Automatically move stop loss to entry price

### User Experience
- **Natural Language Interface**: Chat mode with Wit.ai (English/Spanish)
- **Beautiful UI**: Clean, minimalist interface inspired by Stripe CLI
- **Watch Mode**: Real-time position and order monitoring with auto-refresh
- **Demo Mode**: Complete isolation from live trading for safe testing

### Technical
- **Modular Architecture**: 5 independent modules, each doing one thing well
- **Type Safety**: Strongly-typed with proper error handling
- **BingX Integration**: Full support for BingX perpetual futures
- **Extensible**: Easy to add new brokers, strategies, and NLP providers

## Architecture

trading-cli orchestrates 6 independent Go modules:

```
        ┌──────────────────────┐
        │ trading-common-types │  Shared type definitions
        │      (v0.1.0)        │  Dependencies: None
        └──────────┬───────────┘
                   │
         ┌─────────┴─────────┐
         │                   │
         ↓                   ↓
┌─────────────────┐  ┌─────────────────┐
│  calculator-go  │  │   strategy-go   │  Trading strategies
│  (v0.2.0)       │  │                 │  Dependencies: calculator-go, types
│  Pure math      │  └─────────────────┘
└─────────────────┘
         ↓
┌─────────────────┐  ┌─────────────────┐
│   trading-go    │  │   intent-go     │  NLP processing
│  (v0.1.0)       │  │  (v0.1.0)       │  Dependencies: types
│  Broker API     │  └─────────────────┘
│  Dependencies:  │
│  types          │
└─────────────────┘

         ↓ ↓ ↓ ↓ ↓
┌─────────────────────────────┐
│      trading-cli            │  CLI orchestrator
│                             │  Dependencies: ALL modules above
│  No type conversions! 🎉   │
└─────────────────────────────┘
```

### Why This Architecture?

**Before (Monolithic):**
- 6,000+ lines in one repo
- Tight coupling between components
- Hard to test in isolation
- Difficult to reuse logic

**After (Modular):**
- Each module: 200-500 lines
- Zero circular dependencies
- Each module tested independently
- Modules reusable in other projects

**Key Design Decision:** All modules share types via `trading-common-types`, eliminating the need for type conversions. This keeps modules independent while ensuring type compatibility across the entire system.

## Installation

### From Binary (Recommended)
```bash
go install github.com/agatticelli/trading-cli@latest
```

### From Source
```bash
git clone https://github.com/agatticelli/trading-cli
cd trading-cli
make build
./trading-cli --version
```

### Dependencies

The CLI requires:
- Go 1.21+
- BingX API credentials (demo or production)
- Wit.ai token (optional, for NLP features)

## Quick Start

### 1. Configure Accounts

Create `configs/accounts.yaml`:

```yaml
accounts:
  - name: main
    api_key: your_api_key_here
    secret_key: your_secret_key_here
    broker: bingx
    demo: true  # Start with demo mode!
    enabled: true

  - name: secondary
    api_key: another_api_key
    secret_key: another_secret
    broker: bingx
    demo: true
    enabled: true
```

Get API keys:
- **Demo**: https://bingx.com/en-us/demo/
- **Production**: https://bingx.com/en-us/account/api/

### 2. Test Connection

```bash
# Check balance (demo mode)
./trading-cli --demo balance

# View all accounts
./trading-cli --demo balance --all
```

### 3. Open a Position

```bash
./trading-cli --demo open \
  --symbol BTC-USDT \
  --side long \
  --entry 45000 \
  --sl 44500 \
  --risk 2
```

The CLI will:
1. Calculate position size based on 2% risk
2. Calculate required leverage
3. Set take profit at 2:1 risk-reward ratio
4. Place limit order with TP/SL

### 4. Monitor Positions

```bash
# View positions once
./trading-cli --demo positions

# Watch mode (refresh every 5 seconds)
./trading-cli --demo positions --watch --refresh 5
```

### 5. Natural Language Chat

```bash
./trading-cli --demo chat

> open long BTC at 45000 with stop loss 44500 and risk 2%
> show my positions
> set trailing stop on BTC at 1%
> move ETH to break even
> close 50% of BTC position
```

## Commands Reference

### Account Information

#### balance
View account balance and equity.

```bash
# Single account (default)
./trading-cli --demo balance

# All accounts
./trading-cli --demo balance --all

# Specific account
./trading-cli --demo --account secondary balance

# Watch mode
./trading-cli --demo balance --watch --refresh 10
```

### Position Management

#### positions
View open positions with real-time PnL.

```bash
# View all positions
./trading-cli --demo positions

# Filter by symbol
./trading-cli --demo positions --symbol BTC-USDT

# Watch mode with refresh
./trading-cli --demo positions --watch --refresh 5
```

**Output shows:**
- Symbol, side, size, leverage
- Entry price, mark price
- Unrealized PnL ($ and %)
- Distance to TP/SL

#### orders
View open orders with expected PnL.

```bash
# View all orders
./trading-cli --demo orders

# Filter by symbol
./trading-cli --demo orders --symbol ETH-USDT

# Watch mode
./trading-cli --demo orders --watch
```

### Opening Positions

#### open
Open a new position with risk-based sizing.

```bash
# Basic (auto TP at 2:1 RR)
./trading-cli --demo open \
  --symbol BTC-USDT \
  --side long \
  --entry 45000 \
  --sl 44500 \
  --risk 2

# With custom TP
./trading-cli --demo open \
  --symbol ETH-USDT \
  --side long \
  --entry 3000 \
  --sl 2900 \
  --tp 3200 \
  --risk 1.5

# With custom RR ratio
./trading-cli --demo open \
  --symbol BTC-USDT \
  --side short \
  --entry 45000 \
  --sl 45500 \
  --rr 3 \
  --risk 2

# Market order
./trading-cli --demo open \
  --symbol BTC-USDT \
  --side long \
  --sl 44500 \
  --risk 2 \
  --market
```

**The CLI automatically:**
1. Validates price logic (limit orders don't execute as market)
2. Calculates position size from risk %
3. Calculates required leverage
4. Sets TP at specified RR ratio (default 2:1)
5. Places order with TP/SL atomically

### Closing Positions

#### close
Close position (full or partial).

```bash
# Close full position
./trading-cli --demo close --symbol BTC-USDT

# Close 50%
./trading-cli --demo close --symbol BTC-USDT --percent 50

# Close 25% at specific price
./trading-cli --demo close \
  --symbol ETH-USDT \
  --percent 25 \
  --price 3200

# Market close
./trading-cli --demo close --symbol BTC-USDT --market
```

### Advanced Order Management

#### trail
Set trailing stop loss.

```bash
# Trailing stop at 1% callback
./trading-cli --demo trail \
  --symbol BTC-USDT \
  --activation 46000 \
  --callback 1.0

# Trailing TP (close position with trailing)
./trading-cli --demo trail \
  --symbol ETH-USDT \
  --activation 3200 \
  --callback 0.5
```

**How it works:**
- Order activates when price reaches `--activation`
- Trails price by `--callback` percentage
- Triggers when price retraces by callback amount

#### breakeven
Move stop loss to entry price (lock in zero loss).

```bash
# Move SL to entry
./trading-cli --demo breakeven --symbol BTC-USDT

# All positions to break even
./trading-cli --demo breakeven --all
```

#### cancel
Cancel orders.

```bash
# Cancel all orders for symbol
./trading-cli --demo cancel --symbol BTC-USDT

# Cancel specific order
./trading-cli --demo cancel --symbol BTC-USDT --order-id 123456789
```

### Natural Language Interface

#### chat
Interactive NLP chat mode (requires Wit.ai token).

```bash
export WIT_AI_TOKEN="your-wit-ai-token"
./trading-cli --demo chat
```

**English examples:**
```
> open long BTC at 45000 with stop loss 44500 and risk 2%
> close 50% of ETH position
> set trailing stop on BTC at 1%
> show my positions
> move BTC to break even
> cancel all orders
```

**Spanish examples:**
```
> abrir largo BTC en 45000 con stop loss 44500 y riesgo 2%
> cerrar 50% de posición ETH
> poner trailing stop en BTC al 1%
> mostrar mis posiciones
> mover BTC a break even
> cancelar todas las órdenes
```

## Configuration

### Account Configuration

`configs/accounts.yaml`:

```yaml
accounts:
  # Main trading account
  - name: main
    api_key: ${BINGX_API_KEY}      # Can use env vars
    secret_key: ${BINGX_SECRET_KEY}
    broker: bingx
    demo: true
    enabled: true

  # Secondary account (disabled)
  - name: backup
    api_key: backup_key
    secret_key: backup_secret
    broker: bingx
    demo: false
    enabled: false  # Temporarily disabled
```

### Environment Variables

```bash
# Demo mode (overrides accounts.yaml)
export DEMO_MODE=true

# Wit.ai for NLP features
export WIT_AI_TOKEN="your-token"

# BingX credentials (used in accounts.yaml)
export BINGX_API_KEY="your-key"
export BINGX_SECRET_KEY="your-secret"
```

### Default Settings

Default risk-reward ratio: **2:1**
Default max leverage: **125x**
Default strategy: **risk-ratio**

To customize, edit `internal/executor/executor.go`.

## Shared Types Architecture

All modules use **[trading-common-types](https://github.com/agatticelli/trading-common-types)** for consistency:

```go
// trading-common-types defines all shared types
package types

type Side string
const (
    SideLong  Side = "LONG"
    SideShort Side = "SHORT"
)

type Position struct { ... }
type Order struct { ... }
type OrderRequest struct { ... }
// ... and more
```

**Benefits:**
- ✅ **No type conversions needed** - Data flows directly between modules
- ✅ **Single source of truth** - Type changes in one place
- ✅ **Better maintainability** - Less code, fewer bugs
- ✅ **Scalable** - New projects can reuse the same types

**Module Re-exports:**

Each module re-exports common types for convenience:

```go
// strategy-go/types.go
import "github.com/agatticelli/trading-common-types"

type Side = types.Side  // Re-export for convenience
const (
    SideLong  = types.SideLong
    SideShort = types.SideShort
)
```

This allows using either `types.Side` or `strategy.Side` - both are the same type!

## Error Handling

The CLI provides clear error messages:

```bash
# Invalid price logic
./trading-cli --demo open --symbol BTC-USDT --side long --entry 50000 --sl 49000 --risk 2
# ❌ Error: LONG limit order entry (50000.00) must be below current price (45234.56)

# Insufficient balance
./trading-cli --demo open --symbol BTC-USDT --side long --entry 45000 --sl 44500 --risk 50
# ❌ Error: Insufficient balance. Required: $250.00, Available: $100.00

# Missing parameters
./trading-cli --demo open --symbol BTC-USDT --side long
# ❌ Error: Missing required flags: --entry, --sl, --risk
```

## Examples

### Scenario 1: Swing Trade

```bash
# 1. Open position with 2:1 RR
./trading-cli --demo open \
  --symbol BTC-USDT \
  --side long \
  --entry 45000 \
  --sl 44500 \
  --risk 2

# 2. Monitor in watch mode
./trading-cli --demo positions --watch --refresh 5

# 3. When price moves up, set trailing stop
./trading-cli --demo trail \
  --symbol BTC-USDT \
  --activation 46000 \
  --callback 1

# 4. Or move to break even
./trading-cli --demo breakeven --symbol BTC-USDT
```

### Scenario 2: Quick Scalp

```bash
# 1. Market entry
./trading-cli --demo open \
  --symbol ETH-USDT \
  --side long \
  --sl 2950 \
  --tp 3050 \
  --risk 1 \
  --market

# 2. Monitor closely
./trading-cli --demo positions --watch --refresh 2

# 3. Close 50% at target
./trading-cli --demo close --symbol ETH-USDT --percent 50

# 4. Close remaining at break even
./trading-cli --demo close --symbol ETH-USDT
```

### Scenario 3: Multi-Account

```bash
# Open same position on all accounts
for account in main secondary; do
  ./trading-cli --demo --account $account open \
    --symbol BTC-USDT \
    --side long \
    --entry 45000 \
    --sl 44500 \
    --risk 2
done

# Monitor all accounts
./trading-cli --demo positions --all --watch
```

## Development

### Project Structure

```
trading-cli/
├── cmd/                    # Cobra commands
│   ├── balance.go
│   ├── positions.go
│   ├── open.go
│   ├── close.go
│   ├── chat.go
│   └── ...
├── internal/
│   ├── config/            # Account configuration
│   ├── executor/          # Orchestration + type conversions
│   └── ui/                # Formatters, tables, styles
├── configs/
│   └── accounts.yaml      # Account credentials
├── main.go
└── go.mod
```

### Building

```bash
# Build
make build

# Run tests
make test

# Install locally
make install

# Clean
make clean
```

### Adding a New Command

1. Create `cmd/mycommand.go`
2. Implement Cobra command
3. Use `executor` for orchestration
4. Add to `cmd/root.go`

Example:

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/agatticelli/trading-cli/internal/executor"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Get flags
        symbol, _ := cmd.Flags().GetString("symbol")

        // Use executor
        exec := executor.New(broker, strategy, intentProc, calc)
        return exec.ExecuteMyCommand(ctx, symbol)
    },
}
```

## Troubleshooting

### "No accounts configured"
- Create `configs/accounts.yaml` with at least one account
- Set `enabled: true` for the account

### "API authentication failed"
- Verify API keys are correct
- Check if keys have trading permissions
- For demo mode, use demo API keys from BingX demo site

### "Position size too small"
- Increase risk percentage
- Check minimum order size for the symbol
- Verify you have sufficient balance

### "Invalid symbol"
- Use format: `BTC-USDT`, `ETH-USDT` (not `BTCUSDT` or `BTC/USDT`)
- Check symbol exists on BingX

### Chat mode not working
- Set `WIT_AI_TOKEN` environment variable
- Train your Wit.ai app with trading intents
- Check Wit.ai account is active

## Security

⚠️ **Important Security Notes:**

- **Never commit API keys** to source control
- Use environment variables or secure vaults for credentials
- **Always test with demo mode** before live trading
- Limit API key permissions (trading only, no withdrawals)
- Use different keys for demo and production
- Monitor API key usage regularly

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Documentation

- [MIGRATION_STATUS.md](MIGRATION_STATUS.md) - Migration history and architecture decisions
- [trading-common-types](https://github.com/agatticelli/trading-common-types) - Shared type definitions
- [calculator-go](https://github.com/agatticelli/calculator-go) - Pure math calculations
- [strategy-go](https://github.com/agatticelli/strategy-go) - Trading strategies
- [trading-go](https://github.com/agatticelli/trading-go) - Broker abstraction
- [intent-go](https://github.com/agatticelli/intent-go) - NLP processing

## License

MIT

## Support

- GitHub Issues: https://github.com/agatticelli/trading-cli/issues
- Email: support@example.com (update this)

---

**⚠️ Disclaimer:** This software is for educational purposes. Trading cryptocurrencies carries risk. Use at your own risk.
