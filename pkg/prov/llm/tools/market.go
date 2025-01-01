package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kirill/deriv-teletrader/pkg/core"
	"github.com/tmc/langchaingo/tools"
)

var _ tools.Tool = (*GetPriceTool)(nil)
var _ tools.Tool = (*GetHistoricalDataTool)(nil)

// GetPriceTool is a tool for getting current price of a symbol
type GetPriceTool struct {
	provider core.MarketDataProvider
}

// GetHistoricalDataTool is a tool for getting historical data
type GetHistoricalDataTool struct {
	provider core.MarketDataProvider
}

// NewGetPriceTool creates a new GetPriceTool
func NewGetPriceTool(provider core.MarketDataProvider) *GetPriceTool {
	return &GetPriceTool{provider: provider}
}

// NewGetHistoricalDataTool creates a new GetHistoricalDataTool
func NewGetHistoricalDataTool(provider core.MarketDataProvider) *GetHistoricalDataTool {
	return &GetHistoricalDataTool{provider: provider}
}

// Name implements Tool interface
func (t *GetPriceTool) Name() string {
	return "get_price"
}

// Description implements Tool interface
func (t *GetPriceTool) Description() string {
	return "Get current price for a trading symbol. Input should be a JSON object with 'symbol' field."
}

// Call implements Tool interface
func (t *GetPriceTool) Call(ctx context.Context, input string) (string, error) {
	// Try to parse as JSON first
	var args struct {
		Symbol string `json:"symbol"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		// If not JSON, try to extract symbol from text
		symbol := extractSymbol(input)
		if symbol == "" {
			return "", fmt.Errorf("could not find symbol in input: %s", input)
		}
		args.Symbol = symbol
	}

	price, err := t.provider.GetPrice(ctx, args.Symbol)
	if err != nil {
		return "", fmt.Errorf("failed to get price: %w", err)
	}

	return fmt.Sprintf("Current price for %s: %.2f", args.Symbol, price), nil
}

// Name implements Tool interface
func (t *GetHistoricalDataTool) Name() string {
	return "get_historical_data"
}

// Description implements Tool interface
func (t *GetHistoricalDataTool) Description() string {
	return "Get historical market data for a symbol. Input should be a JSON object with 'symbol', 'interval' (hour/day/week/month), 'style' (ticks/candles), and optional 'count' fields."
}

// Call implements Tool interface
func (t *GetHistoricalDataTool) Call(ctx context.Context, input string) (string, error) {
	// Try to parse as JSON first
	var args struct {
		Symbol   string `json:"symbol"`
		Interval string `json:"interval"`
		Style    string `json:"style"`
		Count    int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		// If not JSON, try to extract parameters from text
		symbol := extractSymbol(input)
		if symbol == "" {
			return "", fmt.Errorf("could not find symbol in input: %s", input)
		}
		args.Symbol = symbol
	}

	if args.Count == 0 {
		args.Count = 10 // default count
	}

	if args.Interval == "" {
		args.Interval = "hour" // default interval
	}

	if args.Style == "" {
		args.Style = "candles" // default style
	}

	req := core.HistoricalDataRequest{
		Symbol:   args.Symbol,
		Interval: core.TimeInterval(args.Interval),
		Style:    core.DataStyle(args.Style),
		Count:    args.Count,
	}

	data, err := t.provider.GetHistoricalData(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get historical data: %w", err)
	}

	// Format the data as a string
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Historical data for %s (%s, %s):\n", args.Symbol, args.Interval, args.Style))
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
}

// extractSymbol tries to find a symbol (like R_50, R_100) in the input text
func extractSymbol(input string) string {
	// Common Deriv symbols
	symbols := []string{"R_10", "R_25", "R_50", "R_75", "R_100"}

	for _, symbol := range symbols {
		if strings.Contains(input, symbol) {
			return symbol
		}
	}
	return ""
}
