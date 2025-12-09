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

		// Initial display
		clearScreen()
		if err := exec.ExecuteGetBalance(cmd.Context()); err != nil {
			return err
		}
		fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", balanceRefresh)

		// Refresh loop - wait happens AFTER each execution
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\n✓ Watch mode stopped")
				return nil
			case <-time.After(time.Duration(balanceRefresh) * time.Second):
				// Capture output in buffer before clearing screen
				output, err := captureBalanceOutput(func() error {
					return exec.ExecuteGetBalance(cmd.Context())
				})

				// Only clear and display if we got output
				if err == nil && output != "" {
					clearScreen()
					fmt.Print(output)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", balanceRefresh)
				} else if err != nil {
					clearScreen()
					fmt.Printf("\nError: %v\n", err)
					fmt.Printf("\n⟳ Refreshing every %ds (Press Ctrl+C to exit)\n", balanceRefresh)
				}
			}
		}
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&balanceWatch, "watch", "w", false, "Continuously refresh display")
	balanceCmd.Flags().IntVarP(&balanceRefresh, "refresh", "r", 30, "Refresh interval in seconds (default: 30)")
}

// captureBalanceOutput captures stdout from a function
func captureBalanceOutput(fn func() error) (string, error) {
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
