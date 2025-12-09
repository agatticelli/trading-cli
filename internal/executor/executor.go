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
	"github.com/agatticelli/trading-cli/internal/ui"
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
		fmt.Println(ui.Account(accountName))

		balance, err := brk.GetBalance(ctx)
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Failed to get balance: %v", err)))
			continue
		}

		fmt.Println(ui.FormatBalance(balance))
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
		fmt.Println(ui.Account(accountName))

		positions, err := brk.GetPositions(ctx, filter)
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Failed to get positions: %v", err)))
			continue
		}

		// Get orders to show TP/SL targets
		orders, err := brk.GetOrders(ctx, &broker.OrderFilter{Symbol: symbol})
		if err != nil {
			// If we can't get orders, still show positions without TP/SL info
			orders = []*broker.Order{}
		}

		// Use table formatter with orders for TP/SL display
		fmt.Println(ui.FormatPositionsTable(positions, orders))
	}

	return nil
}

// ExecuteGetOrders retrieves orders for all accounts
func (e *Executor) ExecuteGetOrders(ctx context.Context, symbol string, verbose bool) error {
	filter := &broker.OrderFilter{}
	if symbol != "" {
		filter.Symbol = symbol
	}

	for accountName, brk := range e.brokers {
		fmt.Println(ui.Account(accountName))

		orders, err := brk.GetOrders(ctx, filter)
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Failed to get orders: %v", err)))
			continue
		}

		// Get positions to calculate expected PnL for TP/SL orders
		positions, err := brk.GetPositions(ctx, &broker.PositionFilter{Symbol: symbol})
		if err != nil {
			// If we can't get positions, still show orders without expected PnL
			positions = []*broker.Position{}
		}

		// Use table formatter with verbose option and positions for PnL calculation
		fmt.Println(ui.FormatOrdersTableWithIDs(orders, positions, verbose))
	}

	return nil
}

// ExecuteCancelOrders cancels orders for all accounts
func (e *Executor) ExecuteCancelOrders(ctx context.Context, symbol string) error {
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		var err error
		if symbol != "" {
			err = brk.CancelAllOrders(ctx, symbol)
			if err != nil {
				fmt.Printf("  ✗ Failed to cancel orders for %s: %v\n", symbol, err)
				continue
			}
			fmt.Printf("  ✓ Canceled all orders for %s\n", symbol)
		} else {
			// Get all positions to cancel orders for each symbol
			positions, err := brk.GetPositions(ctx, &broker.PositionFilter{})
			if err != nil {
				fmt.Printf("  ✗ Failed to get positions: %v\n", err)
				continue
			}

			if len(positions) == 0 {
				fmt.Printf("  No positions with orders to cancel\n")
				continue
			}

			for _, pos := range positions {
				err = brk.CancelAllOrders(ctx, pos.Symbol)
				if err != nil {
					fmt.Printf("  ✗ Failed to cancel orders for %s: %v\n", pos.Symbol, err)
				} else {
					fmt.Printf("  ✓ Canceled orders for %s\n", pos.Symbol)
				}
			}
		}
	}

	return nil
}

// ExecuteClosePosition closes positions for all accounts
func (e *Executor) ExecuteClosePosition(ctx context.Context, symbol string, percentage float64) error {
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		// Get position
		var position *broker.Position
		var err error

		if symbol != "" {
			position, err = brk.GetPosition(ctx, symbol)
		} else {
			// Get all positions and close them
			positions, err := brk.GetPositions(ctx, &broker.PositionFilter{})
			if err != nil {
				fmt.Printf("  ✗ Failed to get positions: %v\n", err)
				continue
			}

			if len(positions) == 0 {
				fmt.Printf("  No positions to close\n")
				continue
			}

			// Close each position
			for _, pos := range positions {
				if err := e.closePosition(ctx, brk, pos, percentage); err != nil {
					fmt.Printf("  ✗ Failed to close %s: %v\n", pos.Symbol, err)
				}
			}
			continue
		}

		if err != nil {
			fmt.Printf("  ✗ Failed to get position: %v\n", err)
			continue
		}

		if position == nil {
			fmt.Printf("  No position found for %s\n", symbol)
			continue
		}

		if err := e.closePosition(ctx, brk, position, percentage); err != nil {
			fmt.Printf("  ✗ Failed to close position: %v\n", err)
		}
	}

	return nil
}

// closePosition closes a single position
func (e *Executor) closePosition(ctx context.Context, brk broker.Broker, pos *broker.Position, percentage float64) error {
	// Calculate size to close
	size := pos.Size
	if percentage > 0 && percentage < 100 {
		size = pos.Size * (percentage / 100)
	}

	// Determine close side (opposite of position side)
	closeSide := broker.SideLong
	if pos.Side == broker.SideLong {
		closeSide = broker.SideShort
	}

	// Place market order to close
	orderReq := &broker.OrderRequest{
		Symbol:     pos.Symbol,
		Side:       closeSide,
		Type:       broker.OrderTypeMarket,
		Size:       size,
		ReduceOnly: true,
	}

	order, err := brk.PlaceOrder(ctx, orderReq)
	if err != nil {
		return err
	}

	if percentage > 0 && percentage < 100 {
		fmt.Printf("  ✓ Closed %.0f%% of %s position (%.4f) | Order: %s\n",
			percentage, pos.Symbol, size, order.ID)
	} else {
		fmt.Printf("  ✓ Closed %s position (%.4f) | Order: %s\n",
			pos.Symbol, size, order.ID)
	}

	return nil
}

// ExecuteTrailingStop sets trailing stop for positions
func (e *Executor) ExecuteTrailingStop(ctx context.Context, symbol string, triggerPrice, callbackRate float64) error {
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		// Get position
		position, err := brk.GetPosition(ctx, symbol)
		if err != nil {
			fmt.Printf("  ✗ Failed to get position: %v\n", err)
			continue
		}

		if position == nil {
			fmt.Printf("  No position found for %s\n", symbol)
			continue
		}

		// Determine side for trailing stop (opposite of position)
		trailSide := broker.SideShort // Close long
		if position.Side == broker.SideShort {
			trailSide = broker.SideLong // Close short
		}

		// Place trailing stop order
		orderReq := &broker.OrderRequest{
			Symbol:     symbol,
			Side:       trailSide,
			Type:       broker.OrderTypeTrailingStop,
			Size:       position.Size,
			ReduceOnly: true,
			Trailing: &broker.TrailingConfig{
				ActivationPrice: triggerPrice,
				CallbackRate:    callbackRate / 100, // Convert percentage to decimal
			},
		}

		order, err := brk.PlaceOrder(ctx, orderReq)
		if err != nil {
			fmt.Printf("  ✗ Failed to place trailing stop: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Trailing stop set for %s\n", symbol)
		fmt.Printf("    Activation: %.2f\n", triggerPrice)
		fmt.Printf("    Callback:   %.2f%%\n", callbackRate)
		fmt.Printf("    Order ID:   %s\n", order.ID)
	}

	return nil
}

// ExecuteBreakEven moves stop loss to entry price
func (e *Executor) ExecuteBreakEven(ctx context.Context, symbol string) error {
	for accountName, brk := range e.brokers {
		fmt.Printf("\n💼 Account: %s\n", accountName)

		// Get position
		position, err := brk.GetPosition(ctx, symbol)
		if err != nil {
			fmt.Printf("  ✗ Failed to get position: %v\n", err)
			continue
		}

		if position == nil {
			fmt.Printf("  No position found for %s\n", symbol)
			continue
		}

		// Cancel existing orders (stop loss)
		if err := brk.CancelAllOrders(ctx, symbol); err != nil {
			fmt.Printf("  ✗ Failed to cancel existing orders: %v\n", err)
			continue
		}

		// Determine side for stop loss (opposite of position)
		stopSide := broker.SideShort
		if position.Side == broker.SideShort {
			stopSide = broker.SideLong
		}

		// Place new stop loss at entry price
		orderReq := &broker.OrderRequest{
			Symbol:     symbol,
			Side:       stopSide,
			Type:       broker.OrderTypeStop,
			Size:       position.Size,
			StopPrice:  position.EntryPrice,
			ReduceOnly: true,
		}

		order, err := brk.PlaceOrder(ctx, orderReq)
		if err != nil {
			fmt.Printf("  ✗ Failed to place break even stop: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Break even set for %s\n", symbol)
		fmt.Printf("    Entry price: %.2f\n", position.EntryPrice)
		fmt.Printf("    Order ID:    %s\n", order.ID)
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
