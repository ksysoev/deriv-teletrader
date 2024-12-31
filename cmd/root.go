package cmd

import (
	"log"

	"github.com/kirill/deriv-teletrader/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config
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

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
}

func initConfig() {
	var err error
	cfg, err = config.InitConfig(cfgFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Bind flags to viper
	viper.BindPFlag("debug", startCmd.Flags().Lookup("debug"))
}
