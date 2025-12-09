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
	ordersSymbol  string
	ordersVerbose bool
	ordersWatch   bool
	ordersRefresh int
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "View open orders",
	Long: `Displays all open orders across enabled accounts

With --verbose flag, shows full order IDs and additional details.
Use --watch to continuously refresh the display at specified intervals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		if !ordersWatch {
			// Single execution
			return exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose)
		}

		// Watch mode - continuous refresh
		if ordersRefresh < 1 {
			return fmt.Errorf("refresh interval must be at least 1 second")
		}

		// Setup signal handling for graceful exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(ordersRefresh) * time.Second)
		defer ticker.Stop()

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", ordersRefresh)

		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-ticker.C:
				clearScreen()
				if err := exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose); err != nil {
					fmt.Printf("\nError: %v\n", err)
				}
				fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", ordersRefresh)
			}
		}
	},
}

func init() {
	ordersCmd.Flags().StringVar(&ordersSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
	ordersCmd.Flags().BoolVarP(&ordersVerbose, "verbose", "v", false, "Show full order IDs and details")
	ordersCmd.Flags().BoolVarP(&ordersWatch, "watch", "w", false, "Continuously refresh display")
	ordersCmd.Flags().IntVarP(&ordersRefresh, "refresh", "r", 5, "Refresh interval in seconds (default: 5)")
}
