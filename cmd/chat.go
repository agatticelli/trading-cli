package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/agatticelli/intent-go"
	"github.com/agatticelli/intent-go/witai"
	"github.com/agatticelli/trading-cli/internal/executor"
	"github.com/agatticelli/trading-cli/internal/ui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive NLP chat mode",
	Long: `Start an interactive chat session with natural language processing.

Powered by Wit.ai, understands trading commands in English and Spanish.

Examples of commands:
  > open long ETH at 3950 with stop loss 3900 and risk 2%
  > show my positions
  > close my ETH position
  > set trailing stop on BTC at 51000 with 0.5% callback
  > what are my open orders?
  > exit

Requires WIT_AI_TOKEN environment variable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exec := getExecutor()

		// Check for Wit.ai token
		token := os.Getenv("WIT_AI_TOKEN")
		if token == "" {
			return fmt.Errorf("WIT_AI_TOKEN environment variable not set")
		}

		// Create Wit.ai processor
		processor, err := witai.New(token)
		if err != nil {
			return fmt.Errorf("failed to create NLP processor: %w", err)
		}

		fmt.Println(ui.HeaderStyle.Render("\n🤖 Trading CLI Chat Mode"))
		fmt.Println(ui.MutedStyle.Render("  Type your trading commands in natural language"))
		fmt.Println(ui.MutedStyle.Render("  Type 'exit' or 'quit' to leave"))
		fmt.Println()

		// Start chat loop
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print(ui.InfoStyle.Render("> "))

			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Check for exit commands
			if input == "exit" || input == "quit" || input == "q" {
				fmt.Println(ui.Success("Goodbye!"))
				break
			}

			// Parse command with Wit.ai
			command, err := processor.ParseCommand(cmd.Context(), input)
			if err != nil {
				fmt.Println(ui.Error(fmt.Sprintf("Failed to parse: %v", err)))
				continue
			}

			// Display parsed intent
			if command.Confidence > 0 {
				fmt.Println(ui.MutedStyle.Render(fmt.Sprintf("  Intent: %s (%.0f%% confidence)",
					command.Intent, command.Confidence*100)))
			}

			// Execute based on intent
			if err := executeNLPCommand(cmd.Context(), exec, command); err != nil {
				fmt.Println(ui.Error(fmt.Sprintf("Execution failed: %v", err)))
			}

			fmt.Println()
		}

		return nil
	},
}

func executeNLPCommand(ctx context.Context, exec *executor.Executor, cmd *intent.NormalizedCommand) error {
	// Validate command
	if !cmd.Valid {
		if len(cmd.Missing) > 0 {
			return fmt.Errorf("missing parameters: %v", cmd.Missing)
		}
		if len(cmd.Errors) > 0 {
			return fmt.Errorf("validation errors: %v", cmd.Errors)
		}
	}

	// Execute based on intent
	switch cmd.Intent {
	case intent.IntentOpenPosition:
		return exec.ExecuteOpenPosition(ctx, cmd, "riskratio")

	case intent.IntentClosePosition:
		symbol := cmd.Symbol
		percentage := 100.0
		return exec.ExecuteClosePosition(ctx, symbol, percentage)

	case intent.IntentViewPositions:
		return exec.ExecuteGetPositions(ctx, cmd.Symbol)

	case intent.IntentViewOrders:
		return exec.ExecuteGetOrders(ctx, cmd.Symbol, false) // Not verbose in chat

	case intent.IntentCancelOrders:
		return exec.ExecuteCancelOrders(ctx, cmd.Symbol)

	case intent.IntentCheckBalance:
		return exec.ExecuteGetBalance(ctx)

	case intent.IntentTrailingStop:
		if cmd.TriggerPrice == nil || cmd.CallbackRate == nil {
			return fmt.Errorf("trailing stop requires trigger price and callback rate")
		}
		return exec.ExecuteTrailingStop(ctx, cmd.Symbol, *cmd.TriggerPrice, *cmd.CallbackRate)

	case intent.IntentBreakEven:
		return exec.ExecuteBreakEven(ctx, cmd.Symbol)

	default:
		return fmt.Errorf("unknown intent: %s", cmd.Intent)
	}
}
