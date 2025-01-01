package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirill/deriv-teletrader/pkg/core"
	"github.com/kirill/deriv-teletrader/pkg/prov/deriv"
	"github.com/kirill/deriv-teletrader/pkg/prov/llm"
	"github.com/kirill/deriv-teletrader/pkg/telegram"
	"github.com/spf13/cobra"
)

// newStartCmd creates and returns the start command
func newStartCmd(cfg **Config) *cobra.Command {
	var debug bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Telegram trading bot",
		Long: `Start the Telegram trading bot that connects to Deriv API
and begins processing trading commands from authorized users.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStartCmd(cmd.Context(), *cfg, debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")

	return cmd
}

// runStartCmd handles the start command execution
func runStartCmd(ctx context.Context, cfg *Config, debug bool) error {
	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Initialize Deriv client
	derivClient, err := deriv.NewClient(&cfg.Deriv)
	if err != nil {
		return fmt.Errorf("failed to create Deriv client: %w", err)
	}

	// Initialize LLM client
	llmClient, err := llm.NewClient(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Initialize core bot
	coreBot, err := core.NewBot(derivClient, llmClient, cfg.Telegram.AllowedUsernames, cfg.Deriv.Symbols)
	if err != nil {
		return err
	}

	// Connect to Deriv API
	if err := derivClient.Connect(ctx); err != nil {
		return err
	}
	defer derivClient.Close()

	// Initialize telegram bot
	bot, err := telegram.NewBot(&cfg.Telegram, coreBot)
	if err != nil {
		return err
	}

	// Start bot
	log.Printf("Starting bot (debug: %v)...\n", debug)
	if err := bot.Start(ctx); err != nil {
		return err
	}

	return nil
}
