package ui

import (
	"fmt"

	"github.com/agatticelli/calculator-go"
	"github.com/agatticelli/trading-go/broker"
)

// Global calculator instance for UI calculations
var calc = calculator.New(125)

// FormatBalance formats a balance display
func FormatBalance(balance *broker.Balance) string {
	// PnL with color
	unrealizedStr := ""
	if balance.UnrealizedPnL > 0 {
		unrealizedStr = SuccessStyle.Render("+" + FormatMoney(balance.UnrealizedPnL))
	} else if balance.UnrealizedPnL < 0 {
		unrealizedStr = ErrorStyle.Render(FormatMoney(balance.UnrealizedPnL))
	} else {
		unrealizedStr = MutedStyle.Render("$0.00")
	}

	data := map[string]string{
		"Asset":          balance.Asset,
		"Total":          FormatMoney(balance.Total),
		"Available":      FormatMoney(balance.Available),
		"In Use":         FormatMoney(balance.InUse),
		"Unrealized PnL": unrealizedStr,
	}

	return RenderSimpleTable(data)
}

// FormatPosition formats a position for compact display
func FormatPosition(pos *broker.Position) string {
	// Icon and side
	sideIcon := IconLong
	sideStyle := LongStyle
	if pos.Side == broker.SideShort {
		sideIcon = IconShort
		sideStyle = ShortStyle
	}

	// PnL color
	pnlStr := ""
	if pos.UnrealizedPnL > 0 {
		pnlStr = SuccessStyle.Render(fmt.Sprintf("+%.2f", pos.UnrealizedPnL))
	} else if pos.UnrealizedPnL < 0 {
		pnlStr = ErrorStyle.Render(fmt.Sprintf("%.2f", pos.UnrealizedPnL))
	} else {
		pnlStr = MutedStyle.Render("0.00")
	}

	return fmt.Sprintf("  %s %s %s | %.4f @ %.2f | PnL: %s | %dx",
		sideIcon,
		sideStyle.Render(pos.Symbol),
		MutedStyle.Render(string(pos.Side)),
		pos.Size,
		pos.EntryPrice,
		pnlStr,
		pos.Leverage,
	)
}

// FormatOrder formats an order for compact display
func FormatOrder(order *broker.Order) string {
	// Type color
	typeStyle := InfoStyle
	if order.Type == broker.OrderTypeMarket {
		typeStyle = BoldStyle
	} else if order.Type == broker.OrderTypeStop || order.Type == broker.OrderTypeTakeProfit {
		typeStyle = WarningStyle
	}

	priceStr := fmt.Sprintf("%.2f", order.Price)
	if order.Price == 0 && order.StopPrice > 0 {
		priceStr = fmt.Sprintf("@ %.2f", order.StopPrice)
	}

	return fmt.Sprintf("  %s %s | %s %s | %.4f %s | %s",
		IconOrder,
		MutedStyle.Render(order.ID),
		BoldStyle.Render(order.Symbol),
		MutedStyle.Render(string(order.Side)),
		order.Size,
		priceStr,
		typeStyle.Render(string(order.Type)),
	)
}

// FormatPositionPlan formats a position plan
func FormatPositionPlan(symbol string, size, entry, sl, tp float64, leverage int, risk, notional float64) string {
	data := map[string]string{
		"Symbol":     symbol,
		"Size":       fmt.Sprintf("%.4f", size),
		"Entry":      FormatMoney(entry),
		"Stop Loss":  FormatMoney(sl),
		"Leverage":   fmt.Sprintf("%dx", leverage),
		"Risk":       FormatMoney(risk),
		"Notional":   FormatMoney(notional),
	}

	if tp > 0 {
		data["Take Profit"] = FormatMoney(tp)
	}

	return "\n" + Box("Position Plan", RenderSimpleTable(data))
}

// FormatPositionsTable formats multiple positions as a table with TP/SL targets
func FormatPositionsTable(positions []*broker.Position, orders []*broker.Order) string {
	if len(positions) == 0 {
		return Info("No open positions")
	}

	// Create order map by symbol and type for quick lookup
	orderMap := make(map[string]map[broker.OrderType]*broker.Order)
	if orders != nil {
		for _, order := range orders {
			if orderMap[order.Symbol] == nil {
				orderMap[order.Symbol] = make(map[broker.OrderType]*broker.Order)
			}
			orderMap[order.Symbol][order.Type] = order
		}
	}

	table := NewTable("Symbol", "Side", "Size", "Entry", "Mark", "PnL", "PnL %", "To TP", "To SL", "Leverage")

	for _, pos := range positions {
		// Side with icon and color
		sideStr := ""
		if pos.Side == broker.SideLong {
			sideStr = LongStyle.Render(IconLong + " LONG")
		} else {
			sideStr = ShortStyle.Render(IconShort + " SHORT")
		}

		// PnL with color
		pnlStr := ""
		if pos.UnrealizedPnL > 0 {
			pnlStr = SuccessStyle.Render("+" + FormatMoney(pos.UnrealizedPnL))
		} else if pos.UnrealizedPnL < 0 {
			pnlStr = ErrorStyle.Render(FormatMoney(pos.UnrealizedPnL))
		} else {
			pnlStr = MutedStyle.Render("$0.00")
		}

		// Calculate PnL percentage using calculator
		pnlPercent := calc.CalculatePnLPercent(pos.Side, pos.EntryPrice, pos.MarkPrice)

		// PnL % with color
		pnlPercentStr := ""
		if pnlPercent > 0 {
			pnlPercentStr = SuccessStyle.Render(fmt.Sprintf("+%.2f%%", pnlPercent))
		} else if pnlPercent < 0 {
			pnlPercentStr = ErrorStyle.Render(fmt.Sprintf("%.2f%%", pnlPercent))
		} else {
			pnlPercentStr = MutedStyle.Render("0.00%")
		}

		// Calculate distance to TP (Take Profit) using calculator
		toTPStr := MutedStyle.Render("-")
		if orderMap[pos.Symbol] != nil && orderMap[pos.Symbol][broker.OrderTypeTakeProfit] != nil {
			tpOrder := orderMap[pos.Symbol][broker.OrderTypeTakeProfit]
			tpPrice := tpOrder.Price
			if tpPrice == 0 {
				tpPrice = tpOrder.StopPrice
			}

			distancePercent := calc.CalculateDistanceToPrice(pos.Side, pos.MarkPrice, tpPrice)

			if distancePercent > 0 {
				toTPStr = SuccessStyle.Render(fmt.Sprintf("+%.2f%%", distancePercent))
			} else {
				toTPStr = ErrorStyle.Render(fmt.Sprintf("%.2f%%", distancePercent))
			}
		}

		// Calculate distance to SL (Stop Loss) using calculator
		toSLStr := MutedStyle.Render("-")
		if orderMap[pos.Symbol] != nil && orderMap[pos.Symbol][broker.OrderTypeStop] != nil {
			slOrder := orderMap[pos.Symbol][broker.OrderTypeStop]
			slPrice := slOrder.Price
			if slPrice == 0 {
				slPrice = slOrder.StopPrice
			}

			distancePercent := calc.CalculateDistanceToPrice(pos.Side, pos.MarkPrice, slPrice)

			if distancePercent < 0 {
				toSLStr = ErrorStyle.Render(fmt.Sprintf("%.2f%%", distancePercent))
			} else {
				toSLStr = WarningStyle.Render(fmt.Sprintf("+%.2f%%", distancePercent))
			}
		}

		table.AddRow(
			BoldStyle.Render(pos.Symbol),
			sideStr,
			fmt.Sprintf("%.4f", pos.Size),
			FormatMoney(pos.EntryPrice),
			FormatMoney(pos.MarkPrice),
			pnlStr,
			pnlPercentStr,
			toTPStr,
			toSLStr,
			fmt.Sprintf("%dx", pos.Leverage),
		)
	}

	return table.Render()
}

// FormatOrdersTable formats multiple orders as a table
func FormatOrdersTable(orders []*broker.Order) string {
	return FormatOrdersTableWithIDs(orders, nil, false)
}

// FormatOrdersTableWithIDs formats orders table with optional full IDs and expected PnL
func FormatOrdersTableWithIDs(orders []*broker.Order, positions []*broker.Position, showFullIDs bool) string {
	if len(orders) == 0 {
		return Info("No open orders")
	}

	// Create position map for quick lookup by symbol
	positionMap := make(map[string]*broker.Position)
	if positions != nil {
		for _, pos := range positions {
			positionMap[pos.Symbol] = pos
		}
	}

	table := NewTable("ID", "Symbol", "Side", "Type", "Size", "Price", "Expected PnL", "Status")

	for _, order := range orders {
		// Side with color
		sideStr := ""
		if order.Side == broker.SideLong {
			sideStr = LongStyle.Render("LONG")
		} else {
			sideStr = ShortStyle.Render("SHORT")
		}

		// Type with color
		typeStr := InfoStyle.Render(string(order.Type))
		if order.Type == broker.OrderTypeMarket {
			typeStr = BoldStyle.Render(string(order.Type))
		}
		// LIMIT, STOP, TAKE_PROFIT all use InfoStyle (blue) for consistency

		// Price
		priceStr := FormatMoney(order.Price)
		if order.Price == 0 && order.StopPrice > 0 {
			priceStr = MutedStyle.Render("@ " + FormatMoney(order.StopPrice))
		}

		// Status with color based on state
		statusStr := ""
		switch order.Status {
		case broker.OrderStatusNew:
			statusStr = InfoStyle.Render(string(order.Status)) // Blue - active and waiting
		case "PENDING":
			statusStr = WarningStyle.Render(string(order.Status)) // Orange - being processed
		case broker.OrderStatusFilled:
			statusStr = SuccessStyle.Render(string(order.Status)) // Green - completed
		case broker.OrderStatusCanceled:
			statusStr = MutedStyle.Render(string(order.Status)) // Gray - canceled
		case broker.OrderStatusRejected:
			statusStr = ErrorStyle.Render(string(order.Status)) // Red - rejected
		default:
			statusStr = MutedStyle.Render(string(order.Status))
		}

		// ID display - full or truncated
		idStr := order.ID
		if !showFullIDs && len(order.ID) > 10 {
			idStr = order.ID[:10] + "..."
		}

		// Calculate expected PnL for orders that could close positions
		expectedPnLStr := MutedStyle.Render("-")
		pos := positionMap[order.Symbol]

		// Only calculate if we have a position and the order could close it
		if pos != nil {
			// Check if order is closing (opposite side or reduce-only)
			isClosing := false

			// TAKE_PROFIT and STOP are always closing orders
			if order.Type == broker.OrderTypeTakeProfit || order.Type == broker.OrderTypeStop {
				isClosing = true
			}

			// For LIMIT orders, check if it's reduce-only or opposite side
			if order.Type == broker.OrderTypeLimit {
				// If reduce-only, it's definitely closing
				if order.ReduceOnly {
					isClosing = true
				} else {
					// Check if order side would close the position
					// LONG position closed by SELL (SHORT) order
					// SHORT position closed by BUY (LONG) order
					if (pos.Side == broker.SideLong && order.Side == broker.SideShort) ||
						(pos.Side == broker.SideShort && order.Side == broker.SideLong) {
						isClosing = true
					}
				}
			}

			if isClosing {
				// Get the execution price (order price or stop price)
				executionPrice := order.Price
				if executionPrice == 0 {
					executionPrice = order.StopPrice
				}

				// Calculate expected PnL using calculator
				pnlNominal, pnlPercent := calc.CalculateExpectedPnL(pos.Side, pos.EntryPrice, executionPrice, order.Size)

				// Format with color
				if pnlNominal > 0 {
					expectedPnLStr = SuccessStyle.Render(fmt.Sprintf("+%s (+%.2f%%)", FormatMoney(pnlNominal), pnlPercent))
				} else if pnlNominal < 0 {
					expectedPnLStr = ErrorStyle.Render(fmt.Sprintf("%s (%.2f%%)", FormatMoney(pnlNominal), pnlPercent))
				} else {
					expectedPnLStr = MutedStyle.Render(fmt.Sprintf("%s (0.00%%)", FormatMoney(0)))
				}
			}
		}

		table.AddRow(
			MutedStyle.Render(idStr),
			BoldStyle.Render(order.Symbol),
			sideStr,
			typeStr,
			fmt.Sprintf("%.4f", order.Size),
			priceStr,
			expectedPnLStr,
			statusStr,
		)
	}

	return table.Render()
}
