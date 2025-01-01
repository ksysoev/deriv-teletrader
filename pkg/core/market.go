package core

import (
	"context"

	"github.com/kirill/deriv-teletrader/pkg/types"
)

// MarketDataProvider defines the interface for fetching market data from different sources
type MarketDataProvider interface {
	// GetHistoricalData retrieves historical market data for a given symbol and time period
	GetHistoricalData(ctx context.Context, req types.HistoricalDataRequest) ([]types.HistoricalDataPoint, error)
	// GetPrice retrieves current price for a symbol
	GetPrice(ctx context.Context, symbol string) (float64, error)
	// GetAvailableSymbols returns a list of available trading symbols
	GetAvailableSymbols(ctx context.Context) ([]string, error)
}

// Re-export types for backward compatibility
type (
	TimeInterval          = types.TimeInterval
	DataStyle             = types.DataStyle
	HistoricalDataPoint   = types.HistoricalDataPoint
	HistoricalDataRequest = types.HistoricalDataRequest
)

// Re-export constants for backward compatibility
const (
	IntervalHour  = types.IntervalHour
	IntervalDay   = types.IntervalDay
	IntervalWeek  = types.IntervalWeek
	IntervalMonth = types.IntervalMonth

	StyleTicks   = types.StyleTicks
	StyleCandles = types.StyleCandles
)
