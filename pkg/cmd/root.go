package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *Config
	rootCmd = &cobra.Command{
		Use:   "deriv-teletrader",
		Short: "A Telegram bot for trading via Deriv API",
		Long: `A Telegram bot that provides the ability to trade on Deriv platform
through their API. It offers various trading commands and real-time
market data access.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext adds all child commands to the root command and sets flags appropriately
// with context support for graceful shutdown.
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
}

func initConfig() {
	var err error
	cfg, err = InitConfig(cfgFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
}
