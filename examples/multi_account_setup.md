# Multi-Account Setup Guide

This guide shows how to configure and use multiple trading accounts with trading-cli.

## Configuration File

The configuration file is located at `~/.trading-cli/config.json` by default.

### Example: Multiple Accounts Configuration

```json
{
  "default_account": "demo_primary",
  "accounts": [
    {
      "name": "demo_primary",
      "exchange": "bingx",
      "api_key": "your-demo-api-key-1",
      "secret_key": "your-demo-secret-key-1",
      "demo": true,
      "max_leverage": 125
    },
    {
      "name": "demo_secondary",
      "exchange": "bingx",
      "api_key": "your-demo-api-key-2",
      "secret_key": "your-demo-secret-key-2",
      "demo": true,
      "max_leverage": 50
    },
    {
      "name": "live_conservative",
      "exchange": "bingx",
      "api_key": "your-live-api-key-1",
      "secret_key": "your-live-secret-key-1",
      "demo": false,
      "max_leverage": 20
    },
    {
      "name": "live_aggressive",
      "exchange": "bingx",
      "api_key": "your-live-api-key-2",
      "secret_key": "your-live-secret-key-2",
      "demo": false,
      "max_leverage": 100
    }
  ]
}
```

## Account Types

### 1. Demo Accounts
- Use for testing strategies
- No real money at risk
- Same API as production
- Great for learning

```bash
# Using demo account (uses default_account)
./trading-cli --demo balance

# Using specific demo account
./trading-cli --account demo_secondary --demo balance
```

### 2. Production Accounts
- Real money trading
- Requires explicit flag
- Always verify commands before executing

```bash
# Using production account (requires explicit flag)
./trading-cli --account live_conservative --live balance

# Opening position on production account
./trading-cli --account live_conservative --live open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 1.0
```

## Usage Patterns

### Pattern 1: Test on Demo, Execute on Live

```bash
# 1. Test strategy on demo account
./trading-cli --account demo_primary --demo open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 2.0

# 2. Monitor demo position
./trading-cli --account demo_primary --demo positions --watch

# 3. If strategy works, execute on live with lower risk
./trading-cli --account live_conservative --live open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 1.0  # Lower risk for live
```

### Pattern 2: Multiple Strategies, Multiple Accounts

```bash
# Conservative strategy on conservative account
./trading-cli --account live_conservative --live open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 1.0 \
  --rr 2.0

# Aggressive strategy on aggressive account
./trading-cli --account live_aggressive --live open \
  --symbol ETH-USDT \
  --side LONG \
  --entry 3000 \
  --sl 2950 \
  --risk 3.0 \
  --rr 3.0
```

### Pattern 3: Watch Multiple Accounts

```bash
# Terminal 1: Watch demo account
./trading-cli --account demo_primary --demo positions --watch --refresh 5

# Terminal 2: Watch live conservative account
./trading-cli --account live_conservative --live positions --watch --refresh 5

# Terminal 3: Watch live aggressive account
./trading-cli --account live_aggressive --live positions --watch --refresh 5
```

## Best Practices

### 1. Account Naming Convention

Use descriptive names that indicate:
- Environment (demo/live)
- Purpose (primary/secondary/testing)
- Strategy (conservative/aggressive)

```
demo_primary
demo_testing
live_conservative
live_aggressive
live_scalping
```

### 2. Leverage Limits

Set appropriate `max_leverage` for each account:

```json
{
  "name": "demo_testing",
  "max_leverage": 125  // Can test high leverage safely
},
{
  "name": "live_conservative",
  "max_leverage": 20   // Conservative for real money
},
{
  "name": "live_aggressive",
  "max_leverage": 50   // Higher risk tolerance
}
```

### 3. Default Account

Set your most-used account as `default_account`:

```json
{
  "default_account": "demo_primary"
}
```

This allows you to omit `--account` flag:

```bash
# Uses default_account
./trading-cli --demo balance

# Explicit account
./trading-cli --account demo_secondary --demo balance
```

### 4. Safety Checks

Always verify:
- Which account you're using
- Demo vs Live flag
- Risk percentage
- Position size

```bash
# Check which account will be used
./trading-cli config show

# Dry-run pattern (check with demo first)
./trading-cli --account demo_primary --demo open [params]  # Test
./trading-cli --account live_conservative --live open [params]  # Execute
```

## Switching Between Accounts

### Method 1: Command-line Flag (Recommended)

```bash
./trading-cli --account demo_primary --demo balance
./trading-cli --account live_conservative --live balance
```

### Method 2: Change Default Account

Edit `~/.trading-cli/config.json`:

```json
{
  "default_account": "live_conservative"
}
```

Then use without `--account` flag:

```bash
./trading-cli --live balance  # Uses live_conservative
```

## Security Considerations

### 1. API Key Permissions

Only grant necessary permissions:
- ✅ Read account balance
- ✅ Read positions
- ✅ Place orders
- ✅ Cancel orders
- ❌ Withdraw funds (NEVER enable)
- ❌ Transfer funds (NEVER enable)

### 2. File Permissions

Protect your config file:

```bash
# Set restrictive permissions
chmod 600 ~/.trading-cli/config.json

# Verify
ls -la ~/.trading-cli/config.json
# Should show: -rw------- (owner read/write only)
```

### 3. Environment Variables (Alternative)

For sensitive environments, use environment variables:

```bash
export TRADING_ACCOUNT_NAME="live_conservative"
export TRADING_API_KEY="your-api-key"
export TRADING_SECRET_KEY="your-secret-key"
export TRADING_DEMO="false"

./trading-cli balance
```

## Common Commands by Account

### Demo Account Commands

```bash
# View all demo accounts
./trading-cli --demo accounts

# Balance check
./trading-cli --account demo_primary --demo balance

# Positions
./trading-cli --account demo_primary --demo positions

# Orders
./trading-cli --account demo_primary --demo orders

# Open position
./trading-cli --account demo_primary --demo open \
  --symbol BTC-USDT --side LONG --entry 45000 --sl 44500 --risk 2.0

# Close position
./trading-cli --account demo_primary --demo close --symbol BTC-USDT
```

### Live Account Commands

```bash
# Balance check (always start here!)
./trading-cli --account live_conservative --live balance

# Check positions before opening new ones
./trading-cli --account live_conservative --live positions

# Open position (with confirmation)
./trading-cli --account live_conservative --live open \
  --symbol BTC-USDT --side LONG --entry 45000 --sl 44500 --risk 1.0 \
  --confirm

# Monitor position
./trading-cli --account live_conservative --live positions --watch

# Close position
./trading-cli --account live_conservative --live close --symbol BTC-USDT
```

## Troubleshooting

### Problem: "Account not found"

**Solution**: Check account name in config:

```bash
# List available accounts
./trading-cli config accounts

# Or check config file
cat ~/.trading-cli/config.json | jq '.accounts[].name'
```

### Problem: "Invalid API credentials"

**Solution**: Verify API keys:

1. Check for extra spaces in config
2. Verify keys are for correct environment (demo vs live)
3. Ensure API key permissions are set correctly

### Problem: "Wrong account used"

**Solution**: Always verify before executing:

```bash
# Check current default account
./trading-cli config show | grep default_account

# Use explicit account flag
./trading-cli --account [name] --demo balance
```

## Example Workflow: Multi-Account Strategy

```bash
#!/bin/bash
# Script: multi-account-strategy.sh

# 1. Check balances on all accounts
echo "=== Checking Balances ==="
./trading-cli --account demo_primary --demo balance
./trading-cli --account live_conservative --live balance

# 2. Open positions on demo first
echo "=== Opening Demo Position ==="
./trading-cli --account demo_primary --demo open \
  --symbol BTC-USDT --side LONG --entry 45000 --sl 44500 --risk 2.0

# 3. Monitor demo for 1 hour
echo "=== Monitoring Demo (press Ctrl+C to stop) ==="
./trading-cli --account demo_primary --demo positions --watch --refresh 60

# 4. If profitable, replicate on live with lower risk
echo "=== Opening Live Position ==="
./trading-cli --account live_conservative --live open \
  --symbol BTC-USDT --side LONG --entry 45000 --sl 44500 --risk 1.0

# 5. Monitor both
echo "=== Final Positions ==="
./trading-cli --account demo_primary --demo positions
./trading-cli --account live_conservative --live positions
```

---

**Remember**: Always test on demo accounts first, and never risk more than you can afford to lose!
