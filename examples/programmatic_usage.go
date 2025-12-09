package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agatticelli/calculator-go"
	"github.com/agatticelli/intent-go"
	"github.com/agatticelli/intent-go/witai"
	"github.com/agatticelli/strategy-go/strategies/riskratio"
	"github.com/agatticelli/trading-cli/internal/config"
	"github.com/agatticelli/trading-cli/internal/executor"
	"github.com/agatticelli/trading-go/bingx"
	"github.com/agatticelli/trading-go/broker"
)

// Example: Using trading-cli components programmatically
// This shows how to integrate trading-cli functionality into your own application

func main() {
	fmt.Println("Trading CLI - Programmatic Usage Example")
	fmt.Println("==========================================\n")

	// 1. Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize components
	client := initializeBroker(cfg)
	calc := calculator.New(125) // Max 125x leverage
	strat := riskratio.New(2.0) // 2:1 risk-reward ratio
	nlpProcessor := initializeNLP()

	// 3. Create executor
	exec := executor.NewExecutor(client, calc, strat)

	ctx := context.Background()

	// Example 1: Get account balance
	fmt.Println("📊 Example 1: Getting account balance...")
	balance, err := client.GetBalance(ctx)
	if err != nil {
		log.Printf("Error getting balance: %v", err)
	} else {
		fmt.Printf("Available: $%.2f\n", balance.Available)
		fmt.Printf("In Use: $%.2f\n", balance.InUse)
		fmt.Printf("Unrealized PnL: $%.2f\n\n", balance.UnrealizedPnL)
	}

	// Example 2: Parse natural language command
	fmt.Println("💬 Example 2: Parsing natural language command...")
	input := "open long BTC at 45000 with stop loss 44500 and risk 2%"
	cmd, err := nlpProcessor.ParseCommand(ctx, input)
	if err != nil {
		log.Printf("Error parsing command: %v", err)
	} else {
		fmt.Printf("Intent: %s\n", cmd.Intent)
		fmt.Printf("Symbol: %s\n", cmd.Symbol)
		fmt.Printf("Side: %s\n", *cmd.Side)
		fmt.Printf("Entry: $%.2f\n", *cmd.EntryPrice)
		fmt.Printf("Stop Loss: $%.2f\n", *cmd.StopLoss)
		fmt.Printf("Risk: %.1f%%\n\n", *cmd.RiskPercent)
	}

	// Example 3: Execute position opening from parsed command
	if cmd != nil && cmd.Valid {
		fmt.Println("🚀 Example 3: Executing parsed command...")
		result, err := exec.ExecuteOpenPosition(ctx, cmd, cfg.DefaultAccountName, balance.Available)
		if err != nil {
			log.Printf("Error executing position: %v", err)
		} else {
			fmt.Printf("Position opened successfully!\n")
			fmt.Printf("Size: %.4f\n", result.Plan.Size)
			fmt.Printf("Leverage: %dx\n", result.Plan.Leverage)
			fmt.Printf("Risk Amount: $%.2f\n\n", result.Plan.RiskAmount)
		}
	}

	// Example 4: Get current positions
	fmt.Println("📈 Example 4: Getting current positions...")
	positions, err := client.GetPositions(ctx, nil)
	if err != nil {
		log.Printf("Error getting positions: %v", err)
	} else {
		fmt.Printf("Found %d open positions\n", len(positions))
		for _, pos := range positions {
			pnlPercent := calc.CalculatePnLPercent(pos.Side, pos.EntryPrice, pos.MarkPrice)
			fmt.Printf("  %s %s: %.4f @ $%.2f (PnL: %.2f%%)\n",
				pos.Symbol, pos.Side, pos.Size, pos.EntryPrice, pnlPercent)
		}
		fmt.Println()
	}

	// Example 5: Calculate position size manually
	fmt.Println("🧮 Example 5: Manual position size calculation...")
	size := calc.CalculateSize(
		balance.Available, // Account balance
		2.0,               // Risk percent
		45000.0,           // Entry price
		44500.0,           // Stop loss
		broker.SideLong,   // Side
	)
	leverage := calc.CalculateLeverage(size, 45000.0, balance.Available, 125)
	fmt.Printf("Calculated size: %.4f BTC\n", size)
	fmt.Printf("Required leverage: %dx\n\n", leverage)

	// Example 6: Place order directly
	fmt.Println("📝 Example 6: Placing order directly...")
	order := &broker.OrderRequest{
		Symbol: "BTC-USDT",
		Side:   broker.SideLong,
		Type:   broker.OrderTypeLimit,
		Size:   0.001,
		Price:  45000.0,
		StopLoss: &broker.StopLossConfig{
			TriggerPrice: 44500.0,
			WorkingType:  broker.WorkingTypeMark,
		},
		TakeProfit: &broker.TakeProfitConfig{
			TriggerPrice: 46000.0,
			WorkingType:  broker.WorkingTypeMark,
		},
	}

	result, err := client.PlaceOrder(ctx, order)
	if err != nil {
		log.Printf("Error placing order: %v", err)
	} else {
		fmt.Printf("Order placed: %s\n", result.ID)
		fmt.Printf("Status: %s\n", result.Status)
	}

	fmt.Println("\n✅ Programmatic usage examples complete!")
}

func loadConfig() (*config.Config, error) {
	configPath := os.Getenv("TRADING_CLI_CONFIG")
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

func initializeBroker(cfg *config.Config) broker.Broker {
	account := cfg.GetAccount(cfg.DefaultAccountName)
	if account == nil {
		log.Fatal("Default account not found")
	}

	// Create BingX client (demo or production based on config)
	client := bingx.NewClient(account.APIKey, account.SecretKey, account.Demo)
	return client
}

func initializeNLP() intent.Processor {
	witToken := os.Getenv("WIT_AI_TOKEN")
	if witToken == "" {
		log.Println("Warning: WIT_AI_TOKEN not set, NLP features will not work")
		return nil
	}

	processor, err := witai.New(witToken)
	if err != nil {
		log.Printf("Warning: Failed to initialize NLP: %v", err)
		return nil
	}

	return processor
}
