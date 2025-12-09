# trading-cli

Minimalist trading CLI for BingX with natural language support.

## Features

- **Multi-Account Support**: Manage multiple trading accounts
- **Natural Language**: Chat interface powered by Wit.ai (English/Spanish)
- **Risk Management**: Automatic position sizing based on risk percentage
- **Advanced Orders**: Multiple TP levels, trailing stops, break even
- **Demo Mode**: Complete isolation from live trading
- **Beautiful UI**: Clean, minimalist interface inspired by Stripe CLI

## Architecture

This CLI is built on top of 3 independent libraries:

- **[trading-go](https://github.com/agatticelli/trading-go)**: Broker abstraction (BingX, extensible to others)
- **[strategy-go](https://github.com/agatticelli/strategy-go)**: Trading strategies and risk management
- **[intent-go](https://github.com/agatticelli/intent-go)**: NLP intent processing (Wit.ai, OpenAI, etc.)

## Installation

```bash
go install github.com/agatticelli/trading-cli@latest
```

Or build from source:

```bash
git clone https://github.com/agatticelli/trading-cli
cd trading-cli
make build
```

## Quick Start

1. **Configure accounts**:
```bash
cp configs/accounts.yaml.example configs/accounts.yaml
# Edit with your API keys
```

2. **Demo mode** (recommended for testing):
```bash
trading-cli --demo balance
trading-cli --demo positions
```

3. **Open a position**:
```bash
trading-cli --demo open \
  --symbol ETH-USDT \
  --side long \
  --entry 3950 \
  --sl 3900 \
  --risk 2
```

4. **Natural language chat**:
```bash
trading-cli --demo chat
> open long ETH at 3950 with stop loss 3900 and risk 2%
> show my positions
> set trailing stop on ETH at 4000 with 0.5% callback
```

## Commands

- `balance` - View account balances
- `positions` - View open positions
- `orders` - View open orders
- `open` - Open a new position
- `close` - Close position (full or partial)
- `trail` - Set trailing take profit
- `cancel` - Cancel orders
- `breakeven` - Move stop loss to entry price
- `chat` - Interactive NLP chat mode

## Configuration

Example `configs/accounts.yaml`:

```yaml
accounts:
  - name: main
    api_key: your_api_key_here
    secret_key: your_secret_key_here
    broker: bingx
    enabled: true
```

## Environment Variables

- `DEMO_MODE=true` - Enable demo mode
- `WIT_AI_TOKEN` - Wit.ai API token for NLP features

## License

MIT
