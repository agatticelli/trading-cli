package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	closeSymbol     string
	closePercentage float64
)

var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "Close positions",
	Long: `Closes positions using market orders. Supports partial closing.

Examples:
  # Close entire ETH-USDT position
  trading-cli --demo close --symbol ETH-USDT

  # Close 50% of BTC-USDT position
  trading-cli --demo close --symbol BTC-USDT --percent 50

  # Close all positions
  trading-cli --demo close`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		// Validate percentage
		if closePercentage < 0 || closePercentage > 100 {
			return fmt.Errorf("percentage must be between 0 and 100")
		}

		return exec.ExecuteClosePosition(cmd.Context(), closeSymbol, closePercentage)
	},
}

func init() {
	closeCmd.Flags().StringVar(&closeSymbol, "symbol", "", "Close specific symbol (default: all)")
	closeCmd.Flags().Float64Var(&closePercentage, "percent", 100, "Percentage to close (1-100)")
}
