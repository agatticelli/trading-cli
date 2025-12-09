package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ordersSymbol string
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "View open orders",
	Long:  `Displays all open orders across enabled accounts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()
		return exec.ExecuteGetOrders(cmd.Context(), ordersSymbol)
	},
}

func init() {
	ordersCmd.Flags().StringVar(&ordersSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
}
