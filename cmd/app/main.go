package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirill/deriv-teletrader/pkg/cmd"
)

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize and execute root command
	rootCmd := cmd.InitCommand()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
