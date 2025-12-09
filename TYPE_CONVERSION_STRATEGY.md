# Shared Types Architecture Strategy

**Last Updated**: December 9, 2024
**Status**: ✅ Implemented

---

## Overview

The trading system uses **shared types** via the `trading-common-types` module to ensure type consistency across all modules while maintaining their independence.

## Architecture Decision

### The Problem

In a modular system with 5+ independent modules (calculator, strategy, trading, intent, CLI), we faced a choice:

**Option A: Independent Types** ❌
- Each module defines its own types (e.g., `calculator.Side`, `broker.Side`, `intent.Side`)
- Requires type conversion functions in the orchestrator
- Every type change needs updates in multiple places
- Conversion logic must be duplicated in every project using these modules

**Option B: Shared Types** ✅ **CHOSEN**
- One `trading-common-types` module defines all shared types
- All modules import and use these types
- No conversion functions needed
- Type changes happen in one place
- New projects inherit type compatibility automatically

### Why We Chose Shared Types

1. **Eliminates Type Conversions**
   ```go
   // Before: Type conversion hell
   func brokerSideFromIntent(side intent.Side) broker.Side {
       if side == intent.SideLong {
           return broker.SideLong
       }
       return broker.SideShort
   }

   // After: Direct usage
   cmd.Side // Already types.Side, no conversion needed!
   ```

2. **Single Source of Truth**
   ```go
   // Add a new field once:
   // trading-common-types/types.go
   type Position struct {
       Symbol string
       Side   Side
       // ... existing fields ...
       IsolatedMargin bool  // ← Add here only
   }

   // All modules instantly have the new field!
   ```

3. **Better Maintainability**
   - **Before**: ~40 lines of conversion functions
   - **After**: 0 lines - functions deleted entirely
   - **Impact**: Less code = fewer bugs

4. **Scalability**
   ```go
   // Future: trading-api project
   import "github.com/agatticelli/trading-common-types"
   import "github.com/agatticelli/trading-go"

   // Works immediately - same types!
   func handleRequest(req types.OrderRequest) {
       broker.PlaceOrder(ctx, &req)  // No conversion needed
   }
   ```

---

## Implementation

### Module Structure

```
trading-common-types/
├── types.go      # Core enums (Side, OrderType, etc.)
├── position.go   # Position, Balance
├── order.go      # Order, OrderRequest, configs
├── strategy.go   # PositionParams, PositionPlan, etc.
├── nlp.go        # NormalizedCommand, TPLevel
└── README.md
```

### Type Categories

#### Core Types (types.go)
```go
type Side string
const (
    SideLong  Side = "LONG"
    SideShort Side = "SHORT"
)

type OrderType string
const (
    OrderTypeMarket  OrderType = "MARKET"
    OrderTypeLimit   OrderType = "LIMIT"
    // ... more
)
```

#### Position Types (position.go)
```go
type Position struct {
    Symbol           string
    Side             Side      // ← Uses common Side type
    Size             float64
    EntryPrice       float64
    MarkPrice        float64
    UnrealizedPnL    float64
    // ... more fields
}
```

#### Strategy Types (strategy.go)
```go
type PositionParams struct {
    Symbol         string
    Side           Side      // ← Same Side type
    EntryPrice     float64
    StopLoss       float64
    AccountBalance float64
    RiskPercent    float64
    MaxLeverage    int
}

type PositionPlan struct {
    Symbol      string
    Side        Side      // ← Same Side type
    Size        float64
    EntryPrice  float64
    Leverage    int
    StopLoss    *StopLossLevel
    TakeProfits []*TakeProfitLevel
    // ... more
}
```

### Module Integration

Each module **re-exports** common types for convenience:

#### calculator-go
```go
// calculator.go
import "github.com/agatticelli/trading-common-types"

// Functions now use types.Side directly
func (c *Calculator) CalculateSize(
    balance, riskPercent, entry, stopLoss float64,
    side types.Side,  // ← Direct import
) float64 {
    // ...
}
```

#### strategy-go
```go
// types.go
import "github.com/agatticelli/trading-common-types"

// Re-export for convenience
type (
    Side           = types.Side
    Position       = types.Position
    PositionParams = types.PositionParams
    PositionPlan   = types.PositionPlan
    // ... more
)

const (
    SideLong  = types.SideLong
    SideShort = types.SideShort
)
```

This allows users to write either:
```go
types.SideLong        // Direct import
strategy.SideLong     // Re-exported - SAME TYPE!
```

#### trading-go
```go
// broker/types.go
import "github.com/agatticelli/trading-common-types"

type (
    Side        = types.Side
    Position    = types.Position
    Order       = types.Order
    OrderRequest = types.OrderRequest
    // ... more
)
```

#### intent-go
```go
// types.go
import "github.com/agatticelli/trading-common-types"

type (
    Intent            = types.Intent
    Side              = types.Side
    NormalizedCommand = types.NormalizedCommand
)
```

#### trading-cli
```go
// internal/executor/executor.go

// Before: Type conversion
calcSide := calculatorSideFromIntent(*cmd.Side)
err := e.calculator.ValidatePriceLogic(calcSide, ...)

// After: Direct usage
err := e.calculator.ValidatePriceLogic(*cmd.Side, ...)  // ✅

// Before: Type conversion
plan, err := strat.CalculatePosition(ctx, strategy.PositionParams{
    Side: strategySideFromIntent(*cmd.Side),  // ❌
    // ...
})

// After: Direct usage
plan, err := strat.CalculatePosition(ctx, strategy.PositionParams{
    Side: *cmd.Side,  // ✅ No conversion!
    // ...
})
```

---

## Benefits Analysis

### Lines of Code Reduction

**Before (Independent Types):**
```go
// trading-cli/internal/executor/executor.go
func brokerSideFromIntent(side intent.Side) broker.Side { ... }      // 7 lines
func strategySideFromIntent(side intent.Side) strategy.Side { ... }  // 7 lines
func calculatorSideFromIntent(side intent.Side) calculator.Side { ... } // 7 lines
func brokerSideFromStrategy(side strategy.Side) broker.Side { ... }  // 7 lines

// trading-cli/internal/ui/formatters.go
func calculatorSideFromBroker(side broker.Side) calculator.Side { ... } // 7 lines

// strategy-go/strategies/riskratio/riskratio.go
func calculatorSideFromStrategy(side strategy.Side) calculator.Side { ... } // 7 lines

// Total: ~42 lines of conversion functions
```

**After (Shared Types):**
```go
// No conversion functions needed!
// Total: 0 lines
// Reduction: 42 lines → 0 lines (100% reduction)
```

### Maintenance Burden

**Scenario: Add new Side value (e.g., `SideNeutral`)**

**Before:**
1. Update `calculator-go/calculator.go` - add `SideNeutral`
2. Update `strategy-go/types.go` - add `SideNeutral`
3. Update `trading-go/broker/types.go` - add `SideNeutral`
4. Update `intent-go/types.go` - add `SideNeutral`
5. Update all conversion functions to handle `SideNeutral`
6. Update tests in all 5 locations

**After:**
1. Update `trading-common-types/types.go` - add `SideNeutral`
2. Done! All modules automatically have it.

### Performance

**Before:**
- Every data transfer requires a function call
- Stack overhead for each conversion
- Compiler cannot inline cross-package conversions

**After:**
- Zero conversion overhead
- Direct memory access
- Compiler can optimize across modules

---

## Design Principles

### 1. No Circular Dependencies

`trading-common-types` has **zero dependencies**:
```go
// trading-common-types/go.mod
module github.com/agatticelli/trading-common-types

go 1.21
// No require statements!
```

This ensures it can be imported by any module without creating circular dependency issues.

### 2. Broker-Agnostic Types

Types are designed to work with any broker:

```go
type OrderRequest struct {
    Symbol      string
    Side        Side
    Type        OrderType
    Size        float64
    Price       float64

    // Optional broker-specific configs
    StopLoss   *StopLossConfig   // Optional
    TakeProfit *TakeProfitConfig // Optional
    Trailing   *TrailingConfig   // Optional
}
```

### 3. Backward Compatibility

Modules re-export types to maintain existing APIs:

```go
// Old code still works:
var side strategy.Side = strategy.SideLong

// New code also works:
var side types.Side = types.SideLong

// They're the SAME type!
```

### 4. Validation-Free

`trading-common-types` only defines types. Validation logic remains in the modules that use them:

- `calculator-go` validates prices and calculations
- `strategy-go` validates strategy parameters
- `trading-go` validates broker requirements
- `intent-go` validates NLP commands

---

## Migration Path

### For Existing Code

**Step 1: Add dependency**
```bash
go get github.com/agatticelli/trading-common-types@v0.1.0
```

**Step 2: Replace imports**
```go
// Before
import "github.com/agatticelli/calculator-go"
side := calculator.SideLong

// After
import "github.com/agatticelli/trading-common-types"
side := types.SideLong
```

**Step 3: Remove conversion functions**
```go
// Before
func convert(side intent.Side) broker.Side {
    if side == intent.SideLong {
        return broker.SideLong
    }
    return broker.SideShort
}

brokerSide := convert(intentSide)

// After
// Just use it directly!
broker.PlaceOrder(ctx, intentSide)
```

### For New Projects

```go
import (
    "github.com/agatticelli/trading-common-types"
    "github.com/agatticelli/trading-go"
    "github.com/agatticelli/strategy-go"
)

func main() {
    // Types work seamlessly across modules
    plan := strategy.Calculate(types.PositionParams{
        Side: types.SideLong,
        // ...
    })

    broker.PlaceOrder(ctx, &types.OrderRequest{
        Side: plan.Side,  // No conversion!
        // ...
    })
}
```

---

## Comparison with Alternatives

### Alternative 1: Interface-Based Abstraction

```go
type Sider interface {
    IsLong() bool
}
```

**Pros:**
- Flexible
- Polymorphic

**Cons:**
- Runtime overhead
- More complex
- Harder to serialize
- Can't use in constants

**Verdict:** ❌ Too complex for simple enums

### Alternative 2: String Constants Only

```go
const (
    SideLong  = "LONG"
    SideShort = "SHORT"
)

func DoSomething(side string) { ... }
```

**Pros:**
- Simple
- No dependencies

**Cons:**
- No type safety
- Typo bugs possible
- IDE autocomplete doesn't help
- Can pass any string

**Verdict:** ❌ Sacrifices too much type safety

### Alternative 3: Shared Types (CHOSEN)

```go
// trading-common-types
type Side string
const (
    SideLong  Side = "LONG"
    SideShort Side = "SHORT"
)
```

**Pros:**
- ✅ Type safe
- ✅ Simple to use
- ✅ No conversion overhead
- ✅ Single source of truth
- ✅ IDE autocomplete works
- ✅ Serializes as string

**Cons:**
- Requires external dependency (but it's pure types, no version conflicts)

**Verdict:** ✅ Best balance of simplicity and safety

---

## Versioning Strategy

### Semantic Versioning

`trading-common-types` follows semver:

- **v0.1.x**: Initial release, add new types/fields
- **v0.2.x**: Minor additions (new optional fields)
- **v1.0.0**: Stable API, backward compatibility guaranteed

### Breaking Changes

**Adding optional fields**: Minor version (v0.1.0 → v0.2.0)
```go
type Position struct {
    Symbol string
    Side   Side
    Size   float64
    // New optional field - minor version bump
    IsolatedMargin *bool  // nil = use default
}
```

**Removing fields**: Major version (v1.0.0 → v2.0.0)
```go
// v1.0.0
type Position struct {
    OldField string  // ← Remove in v2.0.0
}

// v2.0.0
type Position struct {
    // OldField removed - major version bump
}
```

---

## Success Metrics

### Achieved Goals

✅ **Zero conversion functions** - All 42 lines deleted
✅ **Single source of truth** - One file to change types
✅ **Type safety maintained** - Compiler catches mismatches
✅ **Backward compatible** - Existing code still works
✅ **Build performance** - No impact on compile times
✅ **Developer experience** - Simpler to understand and use

---

## Conclusion

The **shared types architecture** via `trading-common-types` successfully eliminates type conversion complexity while maintaining module independence. This approach scales better for future projects and reduces maintenance burden significantly.

**Key Takeaway:** Sometimes the simplest solution (shared types) is better than the "clever" solution (type conversions).

---

## References

- [trading-common-types README](https://github.com/agatticelli/trading-common-types)
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)
- [MIGRATION_STATUS.md](MIGRATION_STATUS.md)
