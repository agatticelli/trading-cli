package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ordersSymbol  string
	ordersVerbose bool
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "View open orders",
	Long: `Displays all open orders across enabled accounts

With --verbose flag, shows full order IDs and additional details.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()
		return exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose)
	},
}

func init() {
	ordersCmd.Flags().StringVar(&ordersSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
	ordersCmd.Flags().BoolVarP(&ordersVerbose, "verbose", "v", false, "Show full order IDs and details")
}
