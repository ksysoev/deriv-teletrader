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

	// Execute root command with context
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
