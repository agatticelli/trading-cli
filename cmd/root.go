package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/agatticelli/trading-cli/internal/config"
	"github.com/agatticelli/trading-cli/internal/executor"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	configPath string
	demoMode   bool

	// Global state
	cfg  *config.Config
	exec *executor.Executor
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "trading-cli",
	Short: "Minimalist trading CLI for BingX with natural language support",
	Long: `A minimalist trading CLI built on modular architecture:

- trading-go: Broker abstraction (BingX, extensible to others)
- strategy-go: Trading strategies and risk management
- intent-go: NLP intent processing (Wit.ai, OpenAI, etc.)

Features:
- Multi-account support
- Natural language chat interface
- Risk-based position sizing
- Advanced orders (TP/SL, trailing stops)
- Demo mode for safe testing`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip initialization for help commands
		if cmd.Name() == "help" || cmd.Parent() == nil {
			return nil
		}

		// Load configuration
		var err error
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Initialize executor
		exec, err = executor.New(cfg, demoMode)
		if err != nil {
			return fmt.Errorf("failed to initialize executor: %w", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "configs/accounts.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVar(&demoMode, "demo", false, "Enable demo/testnet mode")

	// Add subcommands
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(positionsCmd)
	rootCmd.AddCommand(ordersCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(closeCmd)
	rootCmd.AddCommand(cancelCmd)
	rootCmd.AddCommand(trailCmd)
	rootCmd.AddCommand(breakevenCmd)
}

// getExecutor returns the initialized executor or exits
func getExecutor() *executor.Executor {
	if exec == nil {
		fmt.Fprintln(os.Stderr, "Error: executor not initialized")
		os.Exit(1)
	}
	return exec
}
