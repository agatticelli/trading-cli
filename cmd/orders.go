package cmd

import (
	"bytes"
	"fmt"
	"io"
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

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", ordersRefresh)

		// Refresh loop - wait happens AFTER each execution
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-time.After(time.Duration(ordersRefresh) * time.Second):
				// Capture output in buffer before clearing screen
				output, err := captureOutput(func() error {
					return exec.ExecuteGetOrders(cmd.Context(), ordersSymbol, ordersVerbose)
				})

				// Only clear and display if we got output
				if err == nil && output != "" {
					clearScreen()
					fmt.Print(output)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", ordersRefresh)
				} else if err != nil {
					clearScreen()
					fmt.Printf("\nError: %v\n", err)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", ordersRefresh)
				}
			}
		}
	},
}

func init() {
	ordersCmd.Flags().StringVar(&ordersSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
	ordersCmd.Flags().BoolVarP(&ordersVerbose, "verbose", "v", false, "Show full order IDs and details")
	ordersCmd.Flags().BoolVarP(&ordersWatch, "watch", "w", false, "Continuously refresh display")
	ordersCmd.Flags().IntVarP(&ordersRefresh, "refresh", "r", 30, "Refresh interval in seconds (default: 30)")
}

// captureOutput captures stdout from a function
func captureOutput(fn func() error) (string, error) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run function
	err := fn()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), err
}
