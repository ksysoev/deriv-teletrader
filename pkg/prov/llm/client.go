package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kirill/deriv-teletrader/pkg/core"
	"github.com/kirill/deriv-teletrader/pkg/prov/llm/tools"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	lctools "github.com/tmc/langchaingo/tools"
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

// ProcessWithFunctions handles text input with available market data functions
func (c *Client) ProcessWithFunctions(ctx context.Context, input string, provider core.MarketDataProvider, _ []core.LLMFunction) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input text cannot be empty")
	}

	// Create tools
	marketTools := []lctools.Tool{
		tools.NewGetPriceTool(provider),
		tools.NewGetHistoricalDataTool(provider),
	}

	// Create agent executor with more specific instructions
	executor, err := agents.Initialize(
		c.llm,
		marketTools,
		agents.ZeroShotReactDescription,
		agents.WithPromptPrefix(`You are a trading assistant focused on the Deriv trading platform.
You have access to real-time market data through tools. Use these tools to gather data and provide analysis.

When a user asks about a symbol (like R_50, R_100):
1. Use the get_price tool with the exact symbol name to get current price
2. For trend analysis, use get_historical_data with the symbol
3. Analyze the data and explain what it means for trading decisions

Example tool usage:
- To get price: Use get_price with the symbol name (e.g., "R_50")
- To get history: Use get_historical_data with the symbol name

Remember:
- Extract the exact symbol name from user's question (R_10, R_25, R_50, R_75, R_100)
- Always verify data before making suggestions
- Explain your reasoning based on the data
- Keep responses focused on trading information`),
		agents.WithMaxIterations(3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Execute agent with input in the expected format
	agentInput := map[string]any{
		"input": input,
	}

	result, err := executor.Call(ctx, agentInput)
	if err != nil {
		// Try to handle the error gracefully
		if strings.Contains(err.Error(), "could not find symbol") {
			return "I couldn't identify a valid trading symbol in your question. Please specify a symbol like R_50, R_100, etc.", nil
		}
		return "", fmt.Errorf("failed to execute agent: %w", err)
	}

	// Extract output from the result
	output, ok := result["output"].(string)
	if !ok {
		return "I encountered an issue while processing your request. Please try asking in a different way.", nil
	}

	return output, nil
}
