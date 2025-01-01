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
	if input == "" {
		return "", fmt.Errorf("input text cannot be empty")
	}

	// Create prompt with system context and user input
	prompt := "You are a trading assistant focused specifically on the Deriv trading platform. " +
		"Only respond to questions about trading concepts, strategies, market analysis, or the Deriv platform itself. " +
		"If a question is not related to trading or Deriv, politely explain that you can only assist with trading and Deriv-related queries. " +
		"Keep responses clear, concise, and focused on providing accurate trading information.\n\n" +
		"User: " + input + "\n\nAssistant:"

	response, err := c.llm.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to process text: %w", err)
	}

	return response, nil
}
