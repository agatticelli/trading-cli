package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cancelSymbol string
)

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel open orders",
	Long: `Cancels open orders. If no symbol specified, cancels all orders for all positions.

Examples:
  # Cancel all orders for ETH-USDT
  trading-cli --demo cancel --symbol ETH-USDT

  # Cancel all orders for all symbols
  trading-cli --demo cancel`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()
		return exec.ExecuteCancelOrders(cmd.Context(), cancelSymbol)
	},
}

func init() {
	cancelCmd.Flags().StringVar(&cancelSymbol, "symbol", "", "Cancel orders for specific symbol (default: all)")
}
