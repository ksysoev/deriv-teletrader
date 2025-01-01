package core

import "context"

// Available functions for LLM
var MarketDataFunctions = []LLMFunction{
	{
		Name:        "get_price",
		Description: "Get current price for a trading symbol",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "The trading symbol to get price for",
				},
			},
			"required": []string{"symbol"},
		},
	},
	{
		Name:        "get_historical_data",
		Description: "Get historical market data for a symbol",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "The trading symbol to get data for",
				},
				"interval": map[string]interface{}{
					"type":        "string",
					"description": "Time interval (hour, day, week, month)",
					"enum":        []string{"hour", "day", "week", "month"},
				},
				"style": map[string]interface{}{
					"type":        "string",
					"description": "Data style (ticks or candles)",
					"enum":        []string{"ticks", "candles"},
				},
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of data points to return",
					"minimum":     1,
					"maximum":     1000,
				},
			},
			"required": []string{"symbol", "interval"},
		},
	},
}

// LLMFunction represents a function that can be called by the LLM
type LLMFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// LLMFunctionCall represents a function call request from the LLM
type LLMFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// LLMClient defines the interface for LLM operations
type LLMClient interface {
	ProcessText(ctx context.Context, input string) (string, error)
	ProcessWithFunctions(ctx context.Context, input string, provider MarketDataProvider, functions []LLMFunction) (string, error)
}
