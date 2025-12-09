package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - Stripe CLI inspired (muted, professional)
var (
	// Primary colors
	ColorSuccess = lipgloss.Color("#00D06C") // Muted green
	ColorError   = lipgloss.Color("#E25950") // Muted red
	ColorWarning = lipgloss.Color("#F5A623") // Muted orange
	ColorInfo    = lipgloss.Color("#5469D4") // Muted blue
	ColorMuted   = lipgloss.Color("#8898AA") // Gray

	// Position colors
	ColorLong  = lipgloss.Color("#00D06C")
	ColorShort = lipgloss.Color("#E25950")

	// Text colors
	ColorPrimary   = lipgloss.Color("#FFFFFF")
	ColorSecondary = lipgloss.Color("#C4CDD5")
	ColorTertiary  = lipgloss.Color("#8898AA")
)

// Icons - minimalist
const (
	IconSuccess  = "✓"
	IconError    = "✗"
	IconWarning  = "⚠"
	IconInfo     = "ℹ"
	IconAccount  = "💼"
	IconLong     = "↑"
	IconShort    = "↓"
	IconPosition = "▪"
	IconOrder    = "○"
	IconMoney    = "$"
)

// Base styles
var (
	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Info style
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	// Muted style
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Bold style
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// Account header
	AccountStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)

	// Section header
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginTop(1)

	// Key-value pair styles
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorTertiary).
			Width(20)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	// Long position style
	LongStyle = lipgloss.NewStyle().
			Foreground(ColorLong).
			Bold(true)

	// Short position style
	ShortStyle = lipgloss.NewStyle().
			Foreground(ColorShort).
			Bold(true)
)

// Helper functions for formatting

// Success formats a success message
func Success(msg string) string {
	return SuccessStyle.Render(IconSuccess) + " " + msg
}

// Error formats an error message
func Error(msg string) string {
	return ErrorStyle.Render(IconError) + " " + msg
}

// Warning formats a warning message
func Warning(msg string) string {
	return WarningStyle.Render(IconWarning) + " " + msg
}

// Info formats an info message
func Info(msg string) string {
	return InfoStyle.Render(IconInfo) + " " + msg
}

// Account formats an account header
func Account(name string) string {
	return "\n" + AccountStyle.Render(IconAccount+" Account: "+name)
}

// KeyValue formats a key-value pair
func KeyValue(key, value string) string {
	return "  " + KeyStyle.Render(key) + ValueStyle.Render(value)
}

// Money formats a money value with color based on sign
func Money(value float64) string {
	formatted := lipgloss.NewStyle().Foreground(ColorPrimary).Render(FormatMoney(value))
	if value > 0 {
		formatted = SuccessStyle.Render("+") + formatted
	} else if value < 0 {
		formatted = ErrorStyle.Render(formatted)
	}
	return formatted
}

// Percent formats a percentage value
func Percent(value float64) string {
	formatted := formatPercent(value)
	if value > 0 {
		return SuccessStyle.Render("+" + formatted)
	} else if value < 0 {
		return ErrorStyle.Render(formatted)
	}
	return MutedStyle.Render(formatted)
}

// FormatMoney formats a float as money with thousands separators
func FormatMoney(value float64) string {
	// Format with 2 decimal places
	formatted := fmt.Sprintf("%.2f", value)

	// Split into integer and decimal parts
	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Handle negative sign
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}

	// Add thousands separators
	var result strings.Builder
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	// Combine parts
	if negative {
		return fmt.Sprintf("-$%s.%s", result.String(), decPart)
	}
	return fmt.Sprintf("$%s.%s", result.String(), decPart)
}

// formatPercent formats a float as percentage
func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}
