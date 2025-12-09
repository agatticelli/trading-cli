package ui

import (
	"fmt"

	"github.com/agatticelli/trading-go/broker"
)

// FormatBalance formats a balance display
func FormatBalance(balance *broker.Balance) string {
	output := ""
	output += KeyValue("Asset", balance.Asset) + "\n"
	output += KeyValue("Total", fmt.Sprintf("%.2f", balance.Total)) + "\n"
	output += KeyValue("Available", fmt.Sprintf("%.2f", balance.Available)) + "\n"
	output += KeyValue("In Use", fmt.Sprintf("%.2f", balance.InUse)) + "\n"

	// PnL with color
	unrealizedStr := ""
	if balance.UnrealizedPnL > 0 {
		unrealizedStr = SuccessStyle.Render(fmt.Sprintf("+%.2f", balance.UnrealizedPnL))
	} else if balance.UnrealizedPnL < 0 {
		unrealizedStr = ErrorStyle.Render(fmt.Sprintf("%.2f", balance.UnrealizedPnL))
	} else {
		unrealizedStr = fmt.Sprintf("%.2f", balance.UnrealizedPnL)
	}
	output += KeyValue("Unrealized PnL", unrealizedStr) + "\n"

	return output
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
	output := "\n" + HeaderStyle.Render("  Position Plan") + "\n"
	output += KeyValue("Symbol", symbol) + "\n"
	output += KeyValue("Size", fmt.Sprintf("%.4f", size)) + "\n"
	output += KeyValue("Entry", fmt.Sprintf("%.2f", entry)) + "\n"
	if sl > 0 {
		output += KeyValue("Stop Loss", fmt.Sprintf("%.2f", sl)) + "\n"
	}
	if tp > 0 {
		output += KeyValue("Take Profit", fmt.Sprintf("%.2f", tp)) + "\n"
	}
	output += KeyValue("Leverage", fmt.Sprintf("%dx", leverage)) + "\n"
	output += KeyValue("Risk", fmt.Sprintf("$%.2f", risk)) + "\n"
	output += KeyValue("Notional", fmt.Sprintf("$%.2f", notional)) + "\n"

	return output
}
