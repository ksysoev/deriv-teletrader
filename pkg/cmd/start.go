package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
			return runStartCmd(cmd.Context(), &(*cfg).Telegram, debug)
		},
	}

	cmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")

	return cmd
}

// runStartCmd handles the start command execution
func runStartCmd(ctx context.Context, cfg *telegram.Config, debug bool) error {
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

	// Initialize bot
	bot, err := telegram.NewBot(cfg)
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
