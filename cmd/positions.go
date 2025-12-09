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
	positionsSymbol  string
	positionsWatch   bool
	positionsRefresh int
)

var positionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "View open positions",
	Long: `Displays all open positions across enabled accounts

Use --watch to continuously refresh the display at specified intervals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		if !positionsWatch {
			// Single execution
			return exec.ExecuteGetPositions(cmd.Context(), positionsSymbol)
		}

		// Watch mode - continuous refresh
		if positionsRefresh < 1 {
			return fmt.Errorf("refresh interval must be at least 1 second")
		}

		// Setup signal handling for graceful exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(positionsRefresh) * time.Second)
		defer ticker.Stop()

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetPositions(cmd.Context(), positionsSymbol); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", positionsRefresh)

		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-ticker.C:
				clearScreen()
				if err := exec.ExecuteGetPositions(cmd.Context(), positionsSymbol); err != nil {
					fmt.Printf("\nError: %v\n", err)
				}
				fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", positionsRefresh)
			}
		}
	},
}

func init() {
	positionsCmd.Flags().StringVar(&positionsSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
	positionsCmd.Flags().BoolVarP(&positionsWatch, "watch", "w", false, "Continuously refresh display")
	positionsCmd.Flags().IntVarP(&positionsRefresh, "refresh", "r", 5, "Refresh interval in seconds (default: 5)")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
