package cmd

import (
	"github.com/spf13/cobra"
)

var (
	positionsSymbol string
)

var positionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "View open positions",
	Long:  `Displays all open positions across enabled accounts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()
		return exec.ExecuteGetPositions(cmd.Context(), positionsSymbol)
	},
}

func init() {
	positionsCmd.Flags().StringVar(&positionsSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
}
