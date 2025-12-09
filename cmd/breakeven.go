package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	breakevenSymbol string
)

var breakevenCmd = &cobra.Command{
	Use:   "breakeven",
	Short: "Move stop loss to entry price",
	Long: `Moves the stop loss to the entry price (break even point).

This cancels existing stop loss orders and places a new one at entry price,
ensuring the position closes at break even if price retraces.

Examples:
  # Set break even for ETH position
  trading-cli --demo breakeven --symbol ETH-USDT

  # Set break even for BTC position
  trading-cli --demo breakeven --symbol BTC-USDT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		if breakevenSymbol == "" {
			return fmt.Errorf("symbol is required")
		}

		return exec.ExecuteBreakEven(cmd.Context(), breakevenSymbol)
	},
}

func init() {
	breakevenCmd.Flags().StringVar(&breakevenSymbol, "symbol", "", "Trading symbol (required)")
	breakevenCmd.MarkFlagRequired("symbol")
}
