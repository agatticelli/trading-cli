package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table represents a simple table
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new table
func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		headers: headers,
		rows:    [][]string{},
		widths:  widths,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(values ...string) {
	if len(values) != len(t.headers) {
		return // Skip invalid rows
	}

	// Update column widths
	for i, v := range values {
		// Remove ANSI codes for width calculation
		cleanV := stripANSI(v)
		if len(cleanV) > t.widths[i] {
			t.widths[i] = len(cleanV)
		}
	}

	t.rows = append(t.rows, values)
}

// Render renders the table
func (t *Table) Render() string {
	if len(t.rows) == 0 {
		return ""
	}

	var output strings.Builder

	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSecondary).
		PaddingLeft(1).
		PaddingRight(1)

	cellStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1)

	// Border style
	borderStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)

	// Calculate total width
	totalWidth := 1 // Start with left border
	for _, w := range t.widths {
		totalWidth += w + 3 // width + padding (2) + separator (1)
	}

	// Top border
	output.WriteString(borderStyle.Render("┌" + strings.Repeat("─", totalWidth-2) + "┐") + "\n")

	// Headers
	output.WriteString(borderStyle.Render("│"))
	for i, header := range t.headers {
		output.WriteString(headerStyle.Width(t.widths[i]).Render(header))
		if i < len(t.headers)-1 {
			output.WriteString(borderStyle.Render("│"))
		}
	}
	output.WriteString(borderStyle.Render("│") + "\n")

	// Header separator
	output.WriteString(borderStyle.Render("├" + strings.Repeat("─", totalWidth-2) + "┤") + "\n")

	// Rows
	for _, row := range t.rows {
		output.WriteString(borderStyle.Render("│"))
		for i, cell := range row {
			// For styled cells, we need to pad correctly
			cleanCell := stripANSI(cell)
			padding := t.widths[i] - len(cleanCell)
			paddedCell := cellStyle.Render(cell + strings.Repeat(" ", padding))
			output.WriteString(paddedCell)
			if i < len(row)-1 {
				output.WriteString(borderStyle.Render("│"))
			}
		}
		output.WriteString(borderStyle.Render("│") + "\n")
	}

	// Bottom border
	output.WriteString(borderStyle.Render("└" + strings.Repeat("─", totalWidth-2) + "┘") + "\n")

	return output.String()
}

// stripANSI removes ANSI escape codes for length calculation
func stripANSI(s string) string {
	// Simple ANSI stripper (handles most cases)
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// RenderSimpleTable renders a simple key-value table
func RenderSimpleTable(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}

	var output strings.Builder

	// Find max key and value lengths
	maxKeyLen := 0
	maxValueLen := 0
	for key, value := range data {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
		cleanValue := stripANSI(value)
		if len(cleanValue) > maxValueLen {
			maxValueLen = len(cleanValue)
		}
	}

	borderStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	keyStyle := lipgloss.NewStyle().Foreground(ColorTertiary).Bold(true)

	// Calculate exact width: key + spacing + value + padding
	contentWidth := maxKeyLen + 2 + maxValueLen + 2 // key + "  " + value + "  "

	// Top border
	output.WriteString(borderStyle.Render("┌" + strings.Repeat("─", contentWidth) + "┐") + "\n")

	// Rows (need to iterate in consistent order)
	keys := []string{"Asset", "Total", "Available", "In Use", "Unrealized PnL"}
	for _, key := range keys {
		value, exists := data[key]
		if !exists {
			continue
		}

		output.WriteString(borderStyle.Render("│ "))
		output.WriteString(keyStyle.Render(key))

		// Padding after key
		keyPadding := maxKeyLen - len(key)
		output.WriteString(strings.Repeat(" ", keyPadding))
		output.WriteString("  ")

		// Value
		output.WriteString(value)

		// Padding after value to right border
		cleanValue := stripANSI(value)
		valuePadding := maxValueLen - len(cleanValue) + 1
		output.WriteString(strings.Repeat(" ", valuePadding))

		output.WriteString(borderStyle.Render("│") + "\n")
	}

	// Bottom border
	output.WriteString(borderStyle.Render("└" + strings.Repeat("─", contentWidth) + "┘") + "\n")

	return output.String()
}

// Box creates a simple box around content
func Box(title, content string) string {
	lines := strings.Split(content, "\n")
	maxWidth := len(title) + 4

	for _, line := range lines {
		cleanLine := stripANSI(line)
		if len(cleanLine) > maxWidth {
			maxWidth = len(cleanLine)
		}
	}

	borderStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	titleStyle := lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)

	var output strings.Builder

	// Top border with title
	output.WriteString(borderStyle.Render("┌─ "))
	output.WriteString(titleStyle.Render(title))
	output.WriteString(borderStyle.Render(" " + strings.Repeat("─", maxWidth-len(title)-3) + "┐") + "\n")

	// Content
	for _, line := range lines {
		if line == "" {
			continue
		}
		output.WriteString(borderStyle.Render("│ "))
		output.WriteString(line)
		cleanLine := stripANSI(line)
		padding := maxWidth - len(cleanLine) - 1
		if padding > 0 {
			output.WriteString(strings.Repeat(" ", padding))
		}
		output.WriteString(borderStyle.Render("│") + "\n")
	}

	// Bottom border
	output.WriteString(borderStyle.Render("└" + strings.Repeat("─", maxWidth+1) + "┘") + "\n")

	return output.String()
}

// Section creates a section header
func Section(title string) string {
	style := lipgloss.NewStyle().
		Foreground(ColorInfo).
		Bold(true).
		MarginTop(1).
		MarginBottom(0)

	return style.Render("▸ " + title)
}

// Divider creates a visual divider
func Divider() string {
	return MutedStyle.Render(strings.Repeat("─", 60))
}
