package cmd

import (
	"fmt"

	"github.com/agatticelli/intent-go"
	"github.com/spf13/cobra"
)

var (
	openSymbol  string
	openSide    string
	openEntry   float64
	openSL      float64
	openRisk    float64
	openRR      float64
	openTP      float64
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a new position",
	Long: `Opens a new trading position with risk-based position sizing.

Examples:
  # Open long position with 2% risk and 2:1 RR
  trading-cli --demo open --symbol ETH-USDT --side long --entry 3950 --sl 3900 --risk 2 --rr 2

  # Open short position with specific TP
  trading-cli --demo open --symbol BTC-USDT --side short --entry 50000 --sl 51000 --tp 48000 --risk 1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		// Build NormalizedCommand from flags
		command, err := buildNormalizedCommand()
		if err != nil {
			return fmt.Errorf("invalid parameters: %w", err)
		}

		// Validate command
		if !command.Valid {
			if len(command.Missing) > 0 {
				return fmt.Errorf("missing required parameters: %v", command.Missing)
			}
			if len(command.Errors) > 0 {
				return fmt.Errorf("validation errors: %v", command.Errors)
			}
		}

		// Execute with default riskratio strategy
		return exec.ExecuteOpenPosition(cmd.Context(), command, "riskratio")
	},
}

func init() {
	openCmd.Flags().StringVar(&openSymbol, "symbol", "", "Trading symbol (e.g., ETH-USDT)")
	openCmd.Flags().StringVar(&openSide, "side", "", "Position side: long or short")
	openCmd.Flags().Float64Var(&openEntry, "entry", 0, "Entry price")
	openCmd.Flags().Float64Var(&openSL, "sl", 0, "Stop loss price")
	openCmd.Flags().Float64Var(&openRisk, "risk", 0, "Risk percentage (e.g., 2 for 2%)")
	openCmd.Flags().Float64Var(&openRR, "rr", 2.0, "Risk-reward ratio (e.g., 2 for 2:1)")
	openCmd.Flags().Float64Var(&openTP, "tp", 0, "Take profit price (optional, overrides RR)")

	openCmd.MarkFlagRequired("symbol")
	openCmd.MarkFlagRequired("side")
	openCmd.MarkFlagRequired("entry")
	openCmd.MarkFlagRequired("sl")
	openCmd.MarkFlagRequired("risk")
}

func buildNormalizedCommand() (*intent.NormalizedCommand, error) {
	cmd := &intent.NormalizedCommand{
		Intent:      intent.IntentOpenPosition,
		Symbol:      openSymbol,
		EntryPrice:  &openEntry,
		StopLoss:    &openSL,
		RiskPercent: &openRisk,
		RRRatio:     &openRR,
	}

	// Parse side
	switch openSide {
	case "long", "LONG", "largo":
		side := intent.SideLong
		cmd.Side = &side
	case "short", "SHORT", "corto":
		side := intent.SideShort
		cmd.Side = &side
	default:
		return nil, fmt.Errorf("invalid side: %s (use 'long' or 'short')", openSide)
	}

	// If TP specified, use it; otherwise let strategy calculate from RR
	if openTP > 0 {
		cmd.TakeProfit = &openTP
	}

	// Validate
	cmd.Valid = true
	cmd.Missing = []string{}
	cmd.Errors = []string{}

	if cmd.Symbol == "" {
		cmd.Missing = append(cmd.Missing, "symbol")
		cmd.Valid = false
	}
	if cmd.Side == nil {
		cmd.Missing = append(cmd.Missing, "side")
		cmd.Valid = false
	}
	if cmd.EntryPrice == nil || *cmd.EntryPrice <= 0 {
		cmd.Missing = append(cmd.Missing, "entry_price")
		cmd.Valid = false
	}
	if cmd.StopLoss == nil || *cmd.StopLoss <= 0 {
		cmd.Missing = append(cmd.Missing, "stop_loss")
		cmd.Valid = false
	}
	if cmd.RiskPercent == nil || *cmd.RiskPercent <= 0 || *cmd.RiskPercent > 100 {
		cmd.Errors = append(cmd.Errors, "risk must be between 0 and 100")
		cmd.Valid = false
	}

	// Validate price logic
	if cmd.Valid && cmd.Side != nil && cmd.EntryPrice != nil && cmd.StopLoss != nil {
		if *cmd.Side == intent.SideLong && *cmd.StopLoss >= *cmd.EntryPrice {
			cmd.Errors = append(cmd.Errors, "stop loss must be below entry price for LONG positions")
			cmd.Valid = false
		}
		if *cmd.Side == intent.SideShort && *cmd.StopLoss <= *cmd.EntryPrice {
			cmd.Errors = append(cmd.Errors, "stop loss must be above entry price for SHORT positions")
			cmd.Valid = false
		}
	}

	return cmd, nil
}
