package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kirill/deriv-teletrader/pkg/core"
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

// ProcessWithFunctions handles text input with available market data functions
func (c *Client) ProcessWithFunctions(ctx context.Context, input string, provider core.MarketDataProvider, functions []core.LLMFunction) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input text cannot be empty")
	}

	// Create prompt with system context, available functions, and user input
	functionDescriptions := make([]string, len(functions))
	for i, fn := range functions {
		paramsJSON, err := json.MarshalIndent(fn.Parameters, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal function parameters: %w", err)
		}
		functionDescriptions[i] = fmt.Sprintf("Function: %s\nDescription: %s\nParameters: %s\n",
			fn.Name, fn.Description, string(paramsJSON))
	}

	prompt := "You are a trading assistant with access to real-time market data through functions. " +
		"You can use these functions to get market information:\n\n" +
		strings.Join(functionDescriptions, "\n") +
		"\nTo use a function, respond with a JSON object in this format:\n" +
		`{"function": "function_name", "arguments": {"param1": "value1", ...}}` +
		"\n\nUser: " + input + "\n\nAssistant:"

	response, err := c.llm.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to process text: %w", err)
	}

	// Check if response contains a function call
	if strings.Contains(response, `"function":`) {
		var functionCall core.LLMFunctionCall
		if err := json.Unmarshal([]byte(response), &functionCall); err != nil {
			return "", fmt.Errorf("failed to parse function call: %w", err)
		}

		// Execute the function
		result, err := c.executeFunction(ctx, functionCall, provider)
		if err != nil {
			return "", fmt.Errorf("failed to execute function: %w", err)
		}

		// Create follow-up prompt with function result
		followUpPrompt := fmt.Sprintf("%s\n\nFunction result: %s\n\nPlease analyze this data and provide insights:",
			prompt, result)

		response, err = c.llm.Call(ctx, followUpPrompt)
		if err != nil {
			return "", fmt.Errorf("failed to process follow-up: %w", err)
		}
	}

	return response, nil
}

// executeFunction executes a market data function call
func (c *Client) executeFunction(ctx context.Context, call core.LLMFunctionCall, provider core.MarketDataProvider) (string, error) {
	switch call.Name {
	case "get_price":
		symbol, ok := call.Arguments["symbol"].(string)
		if !ok {
			return "", fmt.Errorf("invalid symbol argument")
		}
		price, err := provider.GetPrice(ctx, symbol)
		if err != nil {
			return "", fmt.Errorf("failed to get price: %w", err)
		}
		return fmt.Sprintf("Current price for %s: %.2f", symbol, price), nil

	case "get_historical_data":
		symbol, ok := call.Arguments["symbol"].(string)
		if !ok {
			return "", fmt.Errorf("invalid symbol argument")
		}
		interval, ok := call.Arguments["interval"].(string)
		if !ok {
			interval = "hour" // default interval
		}
		style, ok := call.Arguments["style"].(string)
		if !ok {
			style = "candles" // default style
		}
		count, ok := call.Arguments["count"].(float64)
		if !ok {
			count = 10 // default count
		}

		req := core.HistoricalDataRequest{
			Symbol:   symbol,
			Interval: core.TimeInterval(interval),
			Style:    core.DataStyle(style),
			Count:    int(count),
		}

		data, err := provider.GetHistoricalData(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to get historical data: %w", err)
		}

		// Format the data as a string
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Historical data for %s (%s, %s):\n", symbol, interval, style))
		for _, point := range data {
			if req.Style == core.StyleCandles {
				result.WriteString(fmt.Sprintf("Time: %d, Open: %.2f, High: %.2f, Low: %.2f, Close: %.2f\n",
					point.Timestamp, point.Open, point.High, point.Low, point.Close))
			} else {
				result.WriteString(fmt.Sprintf("Time: %d, Price: %.2f\n",
					point.Timestamp, point.Price))
			}
		}
		return result.String(), nil

	default:
		return "", fmt.Errorf("unknown function: %s", call.Name)
	}
}
