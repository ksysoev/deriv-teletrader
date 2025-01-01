package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// Config holds LLM-specific configuration
type Config struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

type Client struct {
	llm llms.Model
	cfg *Config
}

// NewClient creates a new LLM client using Anthropic's API
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Model == "" {
		cfg.Model = "claude-2" // default model
	}

	// Set environment variables for Anthropic client
	if err := os.Setenv("ANTHROPIC_API_KEY", cfg.APIKey); err != nil {
		return nil, fmt.Errorf("failed to set API key: %w", err)
	}

	llm, err := anthropic.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	return &Client{
		llm: llm,
		cfg: cfg,
	}, nil
}

// ProcessText handles free-form text input and returns a response
func (c *Client) ProcessText(ctx context.Context, input string) (string, error) {
	completion, err := c.llm.Call(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to process text: %w", err)
	}

	return completion, nil
}
