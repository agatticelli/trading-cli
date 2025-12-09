package executor

import (
	"context"
	"fmt"

	"github.com/agatticelli/intent-go"
	"github.com/agatticelli/strategy-go"
	"github.com/agatticelli/strategy-go/strategies/riskratio"
	"github.com/agatticelli/trading-go/bingx"
	"github.com/agatticelli/trading-go/broker"
	"github.com/agatticelli/trading-cli/internal/config"
)

// Executor orchestrates commands across multiple accounts and modules
type Executor struct {
	config     *config.Config
	brokers    map[string]broker.Broker // accountName -> broker
	strategies map[string]strategy.Strategy
	isDemoMode bool
}

// New creates a new executor
func New(cfg *config.Config, isDemoMode bool) (*Executor, error) {
	executor := &Executor{
		config:     cfg,
		brokers:    make(map[string]broker.Broker),
		strategies: make(map[string]strategy.Strategy),
		isDemoMode: isDemoMode,
	}

	// Initialize brokers for each enabled account
	for _, account := range cfg.GetEnabledAccounts() {
		switch account.Broker {
		case "bingx":
			client := bingx.NewClient(account.APIKey, account.SecretKey, isDemoMode)
			executor.brokers[account.Name] = client
		default:
			return nil, fmt.Errorf("unsupported broker: %s", account.Broker)
		}
	}

	// Initialize default strategies
	executor.strategies["riskratio"] = riskratio.New(2.0) // Default 2:1 RR

	return executor, nil
}

// ExecuteOpenPosition opens a position across all accounts
func (e *Executor) ExecuteOpenPosition(ctx context.Context, cmd *intent.NormalizedCommand, strategyName string) error {
	// Get strategy
	strat, ok := e.strategies[strategyName]
	if !ok {
		return fmt.Errorf("strategy not found: %s", strategyName)
	}

	// Execute for each account
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		// 1. Get balance
		balance, err := brk.GetBalance(ctx)
		if err != nil {
			fmt.Printf("  ✗ Failed to get balance: %v\n", err)
			continue
		}

		// 2. Get current price
		currentPrice, err := brk.GetCurrentPrice(ctx, cmd.Symbol)
		if err != nil {
			fmt.Printf("  ✗ Failed to get price: %v\n", err)
			continue
		}

		// 3. Validate price logic
		if err := validatePriceLogic(cmd.Side, *cmd.EntryPrice, *cmd.StopLoss, currentPrice); err != nil {
			fmt.Printf("  ✗ Invalid price logic: %v\n", err)
			continue
		}

		// 4. Calculate position using strategy
		plan, err := strat.CalculatePosition(ctx, strategy.PositionParams{
			Symbol:         cmd.Symbol,
			Side:           brokerSideFromIntent(*cmd.Side),
			EntryPrice:     *cmd.EntryPrice,
			StopLoss:       *cmd.StopLoss,
			AccountBalance: balance.Available,
			RiskPercent:    *cmd.RiskPercent,
			MaxLeverage:    125,
		})
		if err != nil {
			fmt.Printf("  ✗ Position calculation failed: %v\n", err)
			continue
		}

		// 5. Display plan
		displayPositionPlan(plan, balance.Available)

		// 6. Set leverage
		leverageSide := "LONG"
		if plan.Side == broker.SideShort {
			leverageSide = "SHORT"
		}
		if err := brk.SetLeverage(ctx, cmd.Symbol, leverageSide, plan.Leverage); err != nil {
			fmt.Printf("  ✗ Failed to set leverage: %v\n", err)
			continue
		}
		fmt.Printf("  ✓ Leverage set to %dx\n", plan.Leverage)

		// 7. Place order
		orderReq := buildOrderRequest(plan)
		order, err := brk.PlaceOrder(ctx, orderReq)
		if err != nil {
			fmt.Printf("  ✗ Failed to place order: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Order placed: ID %s\n", order.ID)
	}

	return nil
}

// ExecuteGetBalance retrieves balance for all accounts
func (e *Executor) ExecuteGetBalance(ctx context.Context) error {
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		balance, err := brk.GetBalance(ctx)
		if err != nil {
			fmt.Printf("  ✗ Failed to get balance: %v\n", err)
			continue
		}

		fmt.Printf("  Asset:         %s\n", balance.Asset)
		fmt.Printf("  Total:         %.2f\n", balance.Total)
		fmt.Printf("  Available:     %.2f\n", balance.Available)
		fmt.Printf("  In Use:        %.2f\n", balance.InUse)
		fmt.Printf("  Unrealized PnL: %.2f\n", balance.UnrealizedPnL)
	}

	return nil
}

// ExecuteGetPositions retrieves positions for all accounts
func (e *Executor) ExecuteGetPositions(ctx context.Context, symbol string) error {
	filter := &broker.PositionFilter{}
	if symbol != "" {
		filter.Symbol = symbol
	}

	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		positions, err := brk.GetPositions(ctx, filter)
		if err != nil {
			fmt.Printf("  ✗ Failed to get positions: %v\n", err)
			continue
		}

		if len(positions) == 0 {
			fmt.Printf("  No open positions\n")
			continue
		}

		for _, pos := range positions {
			sideIcon := "↑"
			if pos.Side == broker.SideShort {
				sideIcon = "↓"
			}
			fmt.Printf("  %s %s | %.4f @ %.2f | PnL: %.2f | %dx\n",
				sideIcon, pos.Symbol, pos.Size, pos.EntryPrice, pos.UnrealizedPnL, pos.Leverage)
		}
	}

	return nil
}

// Helper functions

func validatePriceLogic(side *intent.Side, entry, stopLoss, currentPrice float64) error {
	if side == nil {
		return fmt.Errorf("side is required")
	}

	// For LONG: SL must be below entry
	if *side == intent.SideLong && stopLoss >= entry {
		return fmt.Errorf("stop loss must be below entry price for LONG positions")
	}

	// For SHORT: SL must be above entry
	if *side == intent.SideShort && stopLoss <= entry {
		return fmt.Errorf("stop loss must be above entry price for SHORT positions")
	}

	// Warn if entry price is far from current price
	priceDiff := ((entry - currentPrice) / currentPrice) * 100
	if priceDiff > 5 || priceDiff < -5 {
		fmt.Printf("  ⚠ Entry price %.2f is %.2f%% away from current price %.2f\n",
			entry, priceDiff, currentPrice)
	}

	return nil
}

func brokerSideFromIntent(side intent.Side) broker.Side {
	if side == intent.SideLong {
		return broker.SideLong
	}
	return broker.SideShort
}

func displayPositionPlan(plan *strategy.PositionPlan, availableBalance float64) {
	fmt.Printf("\n  Position Plan\n")
	fmt.Printf("  Balance:       $%.2f\n", availableBalance)
	fmt.Printf("  Risk Amount:   $%.2f (%.1f%%)\n", plan.RiskAmount, plan.RiskPercent)
	fmt.Printf("  Size:          %.4f\n", plan.Size)
	fmt.Printf("  Entry:         %.2f\n", plan.EntryPrice)
	if plan.StopLoss != nil {
		fmt.Printf("  Stop Loss:     %.2f\n", plan.StopLoss.Price)
	}
	if len(plan.TakeProfits) > 0 {
		fmt.Printf("  Take Profit:   %.2f\n", plan.TakeProfits[0].Price)
	}
	fmt.Printf("  Leverage:      %dx\n", plan.Leverage)
	fmt.Printf("  Notional:      $%.2f\n\n", plan.NotionalValue)
}

func buildOrderRequest(plan *strategy.PositionPlan) *broker.OrderRequest {
	req := &broker.OrderRequest{
		Symbol: plan.Symbol,
		Side:   plan.Side,
		Type:   broker.OrderTypeLimit,
		Size:   plan.Size,
		Price:  plan.EntryPrice,
	}

	// Add stop loss if present
	if plan.StopLoss != nil {
		req.StopLoss = &broker.StopLossConfig{
			TriggerPrice: plan.StopLoss.Price,
			OrderPrice:   0, // Market order
			WorkingType:  broker.WorkingTypeMark,
		}
	}

	// Add take profit if present
	if len(plan.TakeProfits) > 0 {
		req.TakeProfit = &broker.TakeProfitConfig{
			TriggerPrice: plan.TakeProfits[0].Price,
			OrderPrice:   plan.TakeProfits[0].Price, // Limit order
			WorkingType:  broker.WorkingTypeMark,
		}
	}

	return req
}
