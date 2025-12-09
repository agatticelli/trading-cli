# Migration Guide - Trading CLI v1.0

**From Monolithic to Modular Architecture**

This guide helps users and developers understand what changed in the migration from the monolithic architecture to the new 5-module system.

---

## Table of Contents

- [Overview](#overview)
- [What Changed](#what-changed)
- [Breaking Changes](#breaking-changes)
- [Module Map](#module-map)
- [For Users](#for-users)
- [For Developers](#for-developers)
- [Migration Checklist](#migration-checklist)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Overview

### Before (Monolithic)

```
trading-cli/
├── internal/
│   ├── api/         # BingX client (~2000 lines)
│   ├── calculator/  # Position sizing (~500 lines)
│   ├── nlp/         # Wit.ai integration (~800 lines)
│   └── parser/      # Legacy command parsing (~600 lines)
├── cmd/             # CLI commands (~2500 lines)
└── main.go

Total: ~6,400 lines in one repository
```

**Problems:**
- ❌ Tight coupling between components
- ❌ Hard to test individual components
- ❌ Impossible to reuse logic in other projects
- ❌ Difficult to add new brokers or strategies
- ❌ Circular dependencies between packages

### After (Modular)

```
calculator-go/     (~200 lines, zero dependencies)
    ↓
strategy-go/       (~400 lines, depends on calculator-go)
    ↓
trading-go/        (~1000 lines, zero dependencies)
intent-go/         (~500 lines, zero dependencies)
    ↓
trading-cli/       (~2500 lines, orchestrates all modules)

Total: ~4,600 lines across 5 repositories
```

**Benefits:**
- ✅ Each module does one thing well
- ✅ Zero circular dependencies
- ✅ Each module independently testable
- ✅ Modules reusable across projects
- ✅ Easy to swap implementations (e.g., new broker)
- ✅ Cleaner codebase (~30% less code!)

---

## 🆕 NEW: Shared Types Architecture (December 2024)

### trading-common-types Module

We've added a 6th module to eliminate type conversion overhead:

```
trading-common-types/  (~300 lines, zero dependencies)
    ↓
calculator-go, strategy-go, trading-go, intent-go
    ↓
trading-cli
```

**What it provides:**
- Shared type definitions (Side, OrderType, Position, Order, etc.)
- Single source of truth for all types
- Zero type conversion functions needed!

**Before (Type Conversions):**
```go
// Had to convert between modules
func brokerSideFromIntent(side intent.Side) broker.Side {
    if side == intent.SideLong {
        return broker.SideLong
    }
    return broker.SideShort
}
```

**After (Shared Types):**
```go
// All modules use types.Side directly!
import "github.com/agatticelli/trading-common-types"

// No conversion needed
broker.PlaceOrder(ctx, &types.OrderRequest{
    Side: cmd.Side,  // ← Same type everywhere!
})
```

**Impact:**
- ✅ 42 lines of conversion code eliminated
- ✅ Type changes in one place only
- ✅ Better performance (zero conversion overhead)
- ✅ New projects get type compatibility automatically

See [TYPE_CONVERSION_STRATEGY.md](TYPE_CONVERSION_STRATEGY.md) for full details.

---

## What Changed

### Architecture

**Old:** Everything in one repo with internal packages
**New:** 5 independent Go modules with clean interfaces

### Code Organization

| Old Location | New Location | Module |
|-------------|--------------|--------|
| `internal/api/` | `trading-go/bingx/` | trading-go |
| `internal/calculator/` | `calculator-go/` | calculator-go |
| `internal/nlp/` | `intent-go/witai/` | intent-go |
| `internal/strategy/` | `strategy-go/strategies/` | strategy-go |
| `cmd/`, `main.go` | `trading-cli/` | trading-cli |

### Dependencies

**Old:**
```
All code imports from internal/
```

**New:**
```
trading-cli imports:
  - github.com/agatticelli/trading-common-types (shared types)
  - github.com/agatticelli/calculator-go
  - github.com/agatticelli/strategy-go
  - github.com/agatticelli/trading-go
  - github.com/agatticelli/intent-go

All modules import trading-common-types for shared type definitions.
```

---

## Breaking Changes

### For End Users

**✅ No breaking changes!** The CLI commands work exactly the same:

```bash
# These commands work identically
./trading-cli --demo balance
./trading-cli --demo positions
./trading-cli --demo open --symbol BTC-USDT --side long --entry 45000 --sl 44500 --risk 2
./trading-cli --demo chat
```

**Exception:** You need to rebuild the binary after updating:

```bash
# Old way (still works)
git pull
make build

# New way (recommended)
go install github.com/agatticelli/trading-cli@latest
```

### For Developers

**⚠️ Breaking changes if you were importing internal packages:**

#### Import Paths Changed

**Old:**
```go
import (
    "github.com/agatticelli/trading-cli/internal/api"
    "github.com/agatticelli/trading-cli/internal/calculator"
)

client := api.NewBingXClient(key, secret, demo)
size := calculator.CalculateSize(balance, risk, entry, sl)
```

**New:**
```go
import (
    "github.com/agatticelli/trading-go/bingx"
    "github.com/agatticelli/calculator-go"
)

client := bingx.NewClient(key, secret, demo)
calc := calculator.New(125)
size := calc.CalculateSize(balance, risk, entry, sl, calculator.SideLong)
```

#### Type Names Changed

**Old:**
```go
// Types were defined in internal packages
api.Side
api.Position
calculator.Side
```

**New:**
```go
// Each module has its own types
broker.Side       // from trading-go
calculator.Side   // from calculator-go
strategy.Side     // from strategy-go
intent.Side       // from intent-go
```

**Why?** To prevent circular dependencies and keep modules independent.

#### Interface Changes

**Old (calculator):**
```go
// Direct function call
size := calculator.CalculateSize(balance, risk, entry, sl, side)
```

**New (calculator):**
```go
// Method on Calculator struct
calc := calculator.New(maxLeverage)
size := calc.CalculateSize(balance, risk, entry, sl, side)
```

**Old (broker):**
```go
// Concrete BingX client
client := api.NewBingXClient(key, secret, demo)
```

**New (broker):**
```go
// Interface-based
var broker broker.Broker = bingx.NewClient(key, secret, demo)
```

---

## Module Map

Here's where old code went:

### calculator-go

**What it is:** Pure mathematical functions for position sizing, leverage, PnL calculations.

**Migrated from:**
- `internal/calculator/calculator.go`
- `internal/calculator/validation.go`

**No longer includes:**
- ❌ Strategy logic (moved to strategy-go)
- ❌ Broker types (moved to trading-go)

**Now exports:**
```go
type Calculator struct { ... }
func New(maxLeverage int) *Calculator
func (c *Calculator) CalculateSize(...)
func (c *Calculator) CalculateLeverage(...)
func (c *Calculator) CalculateRRTakeProfit(...)
func (c *Calculator) CalculatePnLPercent(...)
func (c *Calculator) ValidateInputs(...)
```

**GitHub:** https://github.com/agatticelli/calculator-go

### strategy-go

**What it is:** Trading strategy implementations and position planning.

**Migrated from:**
- `internal/strategy/riskratio.go` → `strategies/riskratio/riskratio.go`
- Parts of `internal/calculator/` (strategy-specific logic)

**New features:**
- ✅ Strategy interface for pluggable strategies
- ✅ Broker-agnostic types (own `Position`, `OrderRequest`, `Side`)
- ✅ Uses calculator-go for all math

**Now exports:**
```go
type Strategy interface { ... }
type PositionPlan struct { ... }
func (s *RiskRatioStrategy) CalculatePosition(...) (*PositionPlan, error)
```

**GitHub:** https://github.com/agatticelli/strategy-go

### trading-go

**What it is:** Broker abstraction layer with BingX implementation.

**Migrated from:**
- `internal/api/bingx.go` → `bingx/client.go`
- `internal/api/types.go` → `broker/types.go`
- `internal/api/orders.go` → `bingx/orders.go`
- `internal/api/positions.go` → `bingx/positions.go`

**New features:**
- ✅ Broker interface for swappable exchanges
- ✅ Normalized types across brokers
- ✅ Better error handling

**Now exports:**
```go
type Broker interface { ... }
type Position struct { ... }
type Order struct { ... }
func NewClient(apiKey, secretKey string, demo bool) *Client
```

**GitHub:** https://github.com/agatticelli/trading-go

### intent-go

**What it is:** NLP processing for natural language commands.

**Migrated from:**
- `internal/nlp/witai.go` → `witai/witai.go`
- `internal/nlp/types.go` → `types.go`
- `internal/parser/` (legacy code removed)

**New features:**
- ✅ Processor interface for swappable NLP providers
- ✅ Better validation and error messages
- ✅ Multi-language support (English/Spanish)

**Now exports:**
```go
type Processor interface { ... }
type NormalizedCommand struct { ... }
func New(token string) (*WitAIProcessor, error)
func (p *WitAIProcessor) ParseCommand(...) (*NormalizedCommand, error)
```

**GitHub:** https://github.com/agatticelli/intent-go

### trading-cli

**What it is:** CLI orchestrator that ties everything together.

**What remains:**
- `cmd/` - All CLI commands (unchanged)
- `internal/config/` - Account configuration
- `internal/executor/` - Orchestration + type conversions
- `internal/ui/` - Formatters and styles

**What was removed:**
- ❌ `internal/api/` (moved to trading-go)
- ❌ `internal/calculator/` (moved to calculator-go)
- ❌ `internal/nlp/` (moved to intent-go)
- ❌ `internal/parser/` (legacy, removed)

**New responsibilities:**
- ✅ Type conversion between modules
- ✅ Orchestrating module interactions
- ✅ CLI interface and UX

**GitHub:** https://github.com/agatticelli/trading-cli

---

## For Users

### No Action Required! 🎉

If you're just using the CLI, **nothing changes for you**:

- Same commands
- Same flags
- Same behavior
- Same configuration format

### Updating to v1.0

**Option 1: Install from source (recommended)**
```bash
cd trading-cli
git pull
make build
```

**Option 2: Install via go install**
```bash
go install github.com/agatticelli/trading-cli@latest
```

**Option 3: Download binary**
- Check releases: https://github.com/agatticelli/trading-cli/releases

### Configuration

Your `configs/accounts.yaml` works unchanged:

```yaml
accounts:
  - name: main
    api_key: your_key
    secret_key: your_secret
    broker: bingx
    demo: true
    enabled: true
```

---

## For Developers

### Building from Source

**New dependency management:**

```bash
# Clone the CLI
git clone https://github.com/agatticelli/trading-cli
cd trading-cli

# Dependencies are automatically fetched via go.mod
go build -o trading-cli .
```

The CLI uses Go modules with `replace` directives for local development:

```go
// go.mod
replace github.com/agatticelli/calculator-go => ../calculator-go
replace github.com/agatticelli/strategy-go => ../strategy-go
replace github.com/agatticelli/trading-go => ../trading-go
replace github.com/agatticelli/intent-go => ../intent-go
```

### Using Modules in Your Project

**Option 1: Use published modules**
```bash
go get github.com/agatticelli/calculator-go@v0.2.0
go get github.com/agatticelli/strategy-go@latest
go get github.com/agatticelli/trading-go@v0.1.0
go get github.com/agatticelli/intent-go@v0.1.0
```

**Option 2: Local development**
```bash
# Clone all repos
git clone https://github.com/agatticelli/calculator-go
git clone https://github.com/agatticelli/strategy-go
git clone https://github.com/agatticelli/trading-go
git clone https://github.com/agatticelli/intent-go
git clone https://github.com/agatticelli/trading-cli

# Use replace directives in your go.mod
```

### Integrating Modules

**Example: Using calculator-go**
```go
package main

import "github.com/agatticelli/calculator-go"

func main() {
    calc := calculator.New(125)

    size := calc.CalculateSize(
        1000.0,              // balance
        2.0,                 // risk %
        45000.0,             // entry
        44500.0,             // stop loss
        calculator.SideLong,
    )

    leverage := calc.CalculateLeverage(size, 45000.0, 1000.0, 125)
}
```

**Example: Using trading-go**
```go
package main

import (
    "context"
    "github.com/agatticelli/trading-go/bingx"
    "github.com/agatticelli/trading-go/broker"
)

func main() {
    client := bingx.NewClient(apiKey, secretKey, true)

    balance, err := client.GetBalance(context.Background())
    if err != nil {
        panic(err)
    }

    positions, err := client.GetPositions(context.Background(), nil)
}
```

**Example: Using strategy-go**
```go
package main

import (
    "context"
    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/strategy-go/strategies/riskratio"
)

func main() {
    strat := riskratio.New(2.0) // 2:1 RR

    plan, err := strat.CalculatePosition(context.Background(), strategy.PositionParams{
        Symbol:         "BTC-USDT",
        Side:           strategy.SideLong,
        EntryPrice:     45000.0,
        StopLoss:       44500.0,
        AccountBalance: 1000.0,
        RiskPercent:    2.0,
        MaxLeverage:    125,
    })
}
```

**Example: Type conversions in your code**
```go
// Convert between module types
func brokerSideFromIntent(side intent.Side) broker.Side {
    if side == intent.SideLong {
        return broker.SideLong
    }
    return broker.SideShort
}

func calculatorSideFromStrategy(side strategy.Side) calculator.Side {
    if side == strategy.SideLong {
        return calculator.SideLong
    }
    return calculator.SideShort
}
```

---

## Migration Checklist

### For Users
- [ ] Pull latest code: `git pull`
- [ ] Rebuild binary: `make build`
- [ ] Test with demo mode: `./trading-cli --demo balance`
- [ ] Verify commands work as expected
- [ ] ✅ Done! (Told you it was easy)

### For Developers
- [ ] Update import paths in your code
- [ ] Replace `internal/` imports with module imports
- [ ] Update type references (old types → new module types)
- [ ] Add type conversion functions if needed
- [ ] Update tests to use new modules
- [ ] Run `go mod tidy` to clean dependencies
- [ ] Test your integration
- [ ] Update documentation

---

## Troubleshooting

### "Cannot find package internal/api"

**Problem:** Old import paths don't exist.

**Solution:** Update imports:
```go
// Old
import "github.com/agatticelli/trading-cli/internal/api"

// New
import "github.com/agatticelli/trading-go/bingx"
```

### "Type mismatch: broker.Side vs calculator.Side"

**Problem:** Each module defines its own `Side` type.

**Solution:** Add conversion functions:
```go
func calculatorSideFromBroker(side broker.Side) calculator.Side {
    if side == broker.SideLong {
        return calculator.SideLong
    }
    return calculator.SideShort
}
```

See `internal/executor/executor.go` for all conversion functions.

### "go.mod has replace directives"

**Problem:** Local development uses `replace` directives.

**Solution:** This is normal! For local development, keep them. For published code, remove them:

```bash
# Remove all replace directives
go mod edit -dropreplace github.com/agatticelli/calculator-go
go mod edit -dropreplace github.com/agatticelli/strategy-go
# ... etc

# Let Go fetch from GitHub
go mod tidy
```

### "Module not found: calculator-go"

**Problem:** Module not published yet or version mismatch.

**Solution:** Use `replace` directives for local development:

```go
// In your go.mod
replace github.com/agatticelli/calculator-go => ../calculator-go
```

Or wait for the module to be published and tagged.

---

## FAQ

### Why split into modules?

**Reusability:** Each module can be used independently in other projects.

**Testing:** Test each module in isolation without dependencies.

**Maintainability:** Smaller codebases are easier to understand and maintain.

**Flexibility:** Easy to swap implementations (e.g., different broker or NLP provider).

**No Circular Dependencies:** Clean dependency graph prevents import cycles.

### Why does each module define its own `Side` type?

To keep modules **truly independent**. If calculator-go imported `Side` from trading-go, it would depend on trading-go. This would:
- Create circular dependencies (strategy-go → calculator-go → trading-go)
- Make calculator-go unusable without trading-go
- Couple unrelated modules

Instead, trading-cli handles conversions in its adapter layer.

### Can I still use the monolithic version?

Yes, but it's no longer maintained. The last monolithic version is tagged as `v0.9.0`:

```bash
git checkout v0.9.0
make build
```

We strongly recommend upgrading to v1.0 (modular).

### Will the CLI commands change?

**No!** The CLI interface is stable. We're committed to backward compatibility for all commands.

### How do I add a new broker?

1. Implement `broker.Broker` interface in trading-go
2. Create new package: `trading-go/mybroker/`
3. Register in trading-cli's broker factory

See `trading-go/README.md` for details.

### How do I add a new strategy?

1. Implement `strategy.Strategy` interface in strategy-go
2. Create new package: `strategy-go/strategies/mystrategy/`
3. Use it in trading-cli

See `strategy-go/README.md` for details.

### Where's the old `internal/parser/` code?

**Removed!** It was legacy code that's been replaced by intent-go's better NLP processing.

### Do I need to update my account configuration?

**No!** `configs/accounts.yaml` format is unchanged.

### What about my custom scripts?

If your scripts call the CLI commands, they work unchanged:

```bash
# Still works
./trading-cli --demo open --symbol BTC-USDT --side long --entry 45000 --sl 44500 --risk 2
```

If your scripts import Go code, update the import paths (see [For Developers](#for-developers)).

---

## Timeline

- **v0.1.0 - v0.9.0**: Monolithic architecture
- **v0.9.0** (Dec 2024): Last monolithic version
- **v1.0.0** (Dec 2024): Modular architecture
  - calculator-go v0.2.0
  - strategy-go v0.1.0
  - trading-go v0.1.0
  - intent-go v0.1.0

---

## Resources

- **Main Documentation:** [README.md](README.md)
- **Architecture Details:** [MIGRATION_STATUS.md](MIGRATION_STATUS.md)
- **Module Documentation:**
  - [calculator-go](https://github.com/agatticelli/calculator-go)
  - [strategy-go](https://github.com/agatticelli/strategy-go)
  - [trading-go](https://github.com/agatticelli/trading-go)
  - [intent-go](https://github.com/agatticelli/intent-go)

---

## Questions?

- **GitHub Issues:** https://github.com/agatticelli/trading-cli/issues
- **Discussions:** https://github.com/agatticelli/trading-cli/discussions

---

**Happy Trading! 🚀**
