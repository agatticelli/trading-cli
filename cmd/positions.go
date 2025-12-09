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

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetPositions(cmd.Context(), positionsSymbol); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", positionsRefresh)

		// Refresh loop - wait happens AFTER each execution
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-time.After(time.Duration(positionsRefresh) * time.Second):
				// Capture output in buffer before clearing screen
				output, err := captureExecutorOutput(func() error {
					return exec.ExecuteGetPositions(cmd.Context(), positionsSymbol)
				})

				// Only clear and display if we got output
				if err == nil && output != "" {
					clearScreen()
					fmt.Print(output)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", positionsRefresh)
				} else if err != nil {
					clearScreen()
					fmt.Printf("\nError: %v\n", err)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", positionsRefresh)
				}
			}
		}
	},
}

func init() {
	positionsCmd.Flags().StringVar(&positionsSymbol, "symbol", "", "Filter by symbol (e.g., ETH-USDT)")
	positionsCmd.Flags().BoolVarP(&positionsWatch, "watch", "w", false, "Continuously refresh display")
	positionsCmd.Flags().IntVarP(&positionsRefresh, "refresh", "r", 30, "Refresh interval in seconds (default: 30)")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// captureExecutorOutput captures stdout from an executor function
func captureExecutorOutput(fn func() error) (string, error) {
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
