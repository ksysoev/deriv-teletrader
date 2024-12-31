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

var (
	debug bool

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Telegram trading bot",
		Long: `Start the Telegram trading bot that connects to Deriv API
and begins processing trading commands from authorized users.`,
		RunE: runStart,
	}
)

func init() {
	startCmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
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
