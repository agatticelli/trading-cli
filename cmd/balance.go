package cmd

import (
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "View account balances",
	Long:  `Displays balance information for all enabled accounts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()
		return exec.ExecuteGetBalance(cmd.Context())
	},
}
