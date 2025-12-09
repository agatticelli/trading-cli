package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	trailSymbol   string
	trailTrigger  float64
	trailCallback float64
)

var trailCmd = &cobra.Command{
	Use:   "trail",
	Short: "Set trailing stop for position",
	Long: `Sets a trailing stop order for an open position.

The trailing stop will activate when price reaches the trigger price,
then follow the market with the specified callback rate.

Examples:
  # Set trailing stop at 4000 with 0.5% callback
  trading-cli --demo trail --symbol ETH-USDT --trigger 4000 --callback 0.5

  # Tighter trailing with 0.2% callback
  trading-cli --demo trail --symbol BTC-USDT --trigger 51000 --callback 0.2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		// Validate inputs
		if trailSymbol == "" {
			return fmt.Errorf("symbol is required")
		}
		if trailTrigger <= 0 {
			return fmt.Errorf("trigger price must be positive")
		}
		if trailCallback <= 0 || trailCallback > 5 {
			return fmt.Errorf("callback rate must be between 0 and 5%%")
		}

		return exec.ExecuteTrailingStop(cmd.Context(), trailSymbol, trailTrigger, trailCallback)
	},
}

func init() {
	trailCmd.Flags().StringVar(&trailSymbol, "symbol", "", "Trading symbol (required)")
	trailCmd.Flags().Float64Var(&trailTrigger, "trigger", 0, "Activation price (required)")
	trailCmd.Flags().Float64Var(&trailCallback, "callback", 0, "Callback rate percentage (e.g., 0.5 for 0.5%)")

	trailCmd.MarkFlagRequired("symbol")
	trailCmd.MarkFlagRequired("trigger")
	trailCmd.MarkFlagRequired("callback")
}
