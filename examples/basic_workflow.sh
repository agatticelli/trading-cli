#!/bin/bash
# Basic Trading Workflow Example
# This script demonstrates a complete trading workflow using trading-cli

set -e  # Exit on error

# Configuration
ACCOUNT="default"
DEMO_FLAG="--demo"
CLI="./trading-cli"

echo "=========================================="
echo "Trading CLI - Basic Workflow Example"
echo "=========================================="
echo ""

# Step 1: Check account balance
echo "📊 Step 1: Checking account balance..."
$CLI $DEMO_FLAG balance
echo ""

# Step 2: View current positions
echo "📈 Step 2: Viewing current positions..."
$CLI $DEMO_FLAG positions
echo ""

# Step 3: View open orders
echo "📋 Step 3: Viewing open orders..."
$CLI $DEMO_FLAG orders
echo ""

# Step 4: Open a position using NLP chat
echo "💬 Step 4: Opening a position using natural language..."
echo "Command: 'open long BTC at 45000 with stop loss 44500 and risk 2%'"
# Note: This would require interactive input in real usage
# $CLI $DEMO_FLAG chat
echo "(Skipped - requires interactive input)"
echo ""

# Step 5: Open a position using direct command
echo "🚀 Step 5: Opening a position (direct command)..."
$CLI $DEMO_FLAG open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 2.0
echo ""

# Step 6: Check positions again to see the new position
echo "📊 Step 6: Checking positions after opening..."
$CLI $DEMO_FLAG positions
echo ""

# Step 7: Set trailing stop
echo "🎯 Step 7: Setting trailing stop..."
$CLI $DEMO_FLAG trail \
  --symbol BTC-USDT \
  --activation 46000 \
  --callback 1.0
echo ""

# Step 8: Move to break even
echo "⚖️  Step 8: Moving stop loss to break even..."
$CLI $DEMO_FLAG breakeven --symbol BTC-USDT
echo ""

# Step 9: Partial close (50%)
echo "💰 Step 9: Closing 50% of position..."
$CLI $DEMO_FLAG close \
  --symbol BTC-USDT \
  --percentage 50
echo ""

# Step 10: Close remaining position
echo "🏁 Step 10: Closing remaining position..."
$CLI $DEMO_FLAG close --symbol BTC-USDT
echo ""

# Step 11: Cancel any remaining orders
echo "🗑️  Step 11: Canceling remaining orders..."
$CLI $DEMO_FLAG cancel --symbol BTC-USDT
echo ""

# Step 12: Final balance check
echo "💵 Step 12: Final balance check..."
$CLI $DEMO_FLAG balance
echo ""

echo "=========================================="
echo "✅ Workflow complete!"
echo "=========================================="
