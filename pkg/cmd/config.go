package cmd

import (
	"fmt"
	"strings"

	"github.com/kirill/deriv-teletrader/pkg/deriv"
	"github.com/kirill/deriv-teletrader/pkg/telegram"
	"github.com/spf13/viper"
)

type Config struct {
	// Telegram settings
	Telegram telegram.Config `mapstructure:"telegram"`

	// Deriv API settings
	Deriv deriv.Config `mapstructure:"deriv"`
}

// InitConfig initializes the configuration using Viper
func InitConfig(cfgFile string) (*Config, error) {
	// Set default values
	setDefaults()

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read environment variables
	viper.SetEnvPrefix("TELETRADER")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is ok as we can use env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("deriv.endpoint", "wss://ws.binaryws.com/websockets/v3")
	viper.SetDefault("deriv.symbols", []string{"R_10", "R_25", "R_50", "R_75", "R_100"})
	viper.SetDefault("debug", false)
}

func (c *Config) validate() error {
	if c.Telegram.Token == "" {
		return fmt.Errorf("telegram.token is required")
	}
	if len(c.Telegram.AllowedUsernames) == 0 {
		return fmt.Errorf("telegram.allowed_usernames is required")
	}
	if c.Deriv.AppID == "" {
		return fmt.Errorf("deriv.app_id is required")
	}
	if c.Deriv.APIToken == "" {
		return fmt.Errorf("deriv.api_token is required")
	}
	return nil
}
