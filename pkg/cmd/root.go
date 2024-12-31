package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// InitCommand initializes and returns the root command
func InitCommand() *cobra.Command {
	var cfgFile string
	var cfg *Config

	rootCmd := &cobra.Command{
		Use:   "deriv-teletrader",
		Short: "A Telegram bot for trading via Deriv API",
		Long: `A Telegram bot that provides the ability to trade on Deriv platform
through their API. It offers various trading commands and real-time
market data access.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = InitConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}
			return nil
		},
	}

	// Add flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Add commands
	rootCmd.AddCommand(newStartCmd(&cfg))

	return rootCmd
}
