package ui

import (
	"fmt"

	"github.com/agatticelli/trading-go/broker"
)

// FormatBalance formats a balance display
func FormatBalance(balance *broker.Balance) string {
	// PnL with color
	unrealizedStr := ""
	if balance.UnrealizedPnL > 0 {
		unrealizedStr = SuccessStyle.Render(fmt.Sprintf("+$%.2f", balance.UnrealizedPnL))
	} else if balance.UnrealizedPnL < 0 {
		unrealizedStr = ErrorStyle.Render(fmt.Sprintf("$%.2f", balance.UnrealizedPnL))
	} else {
		unrealizedStr = MutedStyle.Render("$0.00")
	}

	data := map[string]string{
		"Asset":         balance.Asset,
		"Total":         fmt.Sprintf("$%.2f", balance.Total),
		"Available":     fmt.Sprintf("$%.2f", balance.Available),
		"In Use":        fmt.Sprintf("$%.2f", balance.InUse),
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
		"Entry":      fmt.Sprintf("$%.2f", entry),
		"Stop Loss":  fmt.Sprintf("$%.2f", sl),
		"Leverage":   fmt.Sprintf("%dx", leverage),
		"Risk":       fmt.Sprintf("$%.2f", risk),
		"Notional":   fmt.Sprintf("$%.2f", notional),
	}

	if tp > 0 {
		data["Take Profit"] = fmt.Sprintf("$%.2f", tp)
	}

	return "\n" + Box("Position Plan", RenderSimpleTable(data))
}

// FormatPositionsTable formats multiple positions as a table
func FormatPositionsTable(positions []*broker.Position) string {
	if len(positions) == 0 {
		return Info("No open positions")
	}

	table := NewTable("Symbol", "Side", "Size", "Entry", "Mark", "PnL", "Leverage")

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
			pnlStr = SuccessStyle.Render(fmt.Sprintf("+$%.2f", pos.UnrealizedPnL))
		} else if pos.UnrealizedPnL < 0 {
			pnlStr = ErrorStyle.Render(fmt.Sprintf("$%.2f", pos.UnrealizedPnL))
		} else {
			pnlStr = MutedStyle.Render("$0.00")
		}

		table.AddRow(
			BoldStyle.Render(pos.Symbol),
			sideStr,
			fmt.Sprintf("%.4f", pos.Size),
			fmt.Sprintf("$%.2f", pos.EntryPrice),
			fmt.Sprintf("$%.2f", pos.MarkPrice),
			pnlStr,
			fmt.Sprintf("%dx", pos.Leverage),
		)
	}

	return table.Render()
}

// FormatOrdersTable formats multiple orders as a table
func FormatOrdersTable(orders []*broker.Order) string {
	return FormatOrdersTableWithIDs(orders, false)
}

// FormatOrdersTableWithIDs formats orders table with optional full IDs
func FormatOrdersTableWithIDs(orders []*broker.Order, showFullIDs bool) string {
	if len(orders) == 0 {
		return Info("No open orders")
	}

	table := NewTable("ID", "Symbol", "Side", "Type", "Size", "Price", "Status")

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
		} else if order.Type == broker.OrderTypeStop {
			typeStr = WarningStyle.Render(string(order.Type))
		}

		// Price
		priceStr := fmt.Sprintf("$%.2f", order.Price)
		if order.Price == 0 && order.StopPrice > 0 {
			priceStr = MutedStyle.Render(fmt.Sprintf("@ $%.2f", order.StopPrice))
		}

		// Status
		statusStr := MutedStyle.Render(string(order.Status))

		// ID display - full or truncated
		idStr := order.ID
		if !showFullIDs && len(order.ID) > 10 {
			idStr = order.ID[:10] + "..."
		}

		table.AddRow(
			MutedStyle.Render(idStr),
			BoldStyle.Render(order.Symbol),
			sideStr,
			typeStr,
			fmt.Sprintf("%.4f", order.Size),
			priceStr,
			statusStr,
		)
	}

	output := table.Render()

	// Add copyable IDs section if there are orders
	if showFullIDs && len(orders) > 0 {
		output += "\n" + Section("Order IDs (copy-paste ready)") + "\n"
		for _, order := range orders {
			output += MutedStyle.Render(fmt.Sprintf("  %s: %s\n", order.Symbol, order.ID))
		}
	}

	return output
}
