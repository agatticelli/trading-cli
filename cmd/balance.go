package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	balanceWatch   bool
	balanceRefresh int
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "View account balances",
	Long: `Displays balance information for all enabled accounts

Use --watch to continuously refresh the display at specified intervals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		if !balanceWatch {
			// Single execution
			return exec.ExecuteGetBalance(cmd.Context())
		}

		// Watch mode - continuous refresh
		if balanceRefresh < 1 {
			return fmt.Errorf("refresh interval must be at least 1 second")
		}

		// Setup signal handling for graceful exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(balanceRefresh) * time.Second)
		defer ticker.Stop()

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetBalance(cmd.Context()); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", balanceRefresh)

		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-ticker.C:
				clearScreen()
				if err := exec.ExecuteGetBalance(cmd.Context()); err != nil {
					fmt.Printf("\nError: %v\n", err)
				}
				fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", balanceRefresh)
			}
		}
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&balanceWatch, "watch", "w", false, "Continuously refresh display")
	balanceCmd.Flags().IntVarP(&balanceRefresh, "refresh", "r", 5, "Refresh interval in seconds (default: 5)")
}
