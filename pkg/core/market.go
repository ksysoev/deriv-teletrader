package core

// TimeInterval represents different time intervals for historical data
type TimeInterval string

const (
	IntervalHour  TimeInterval = "hour"
	IntervalDay   TimeInterval = "day"
	IntervalWeek  TimeInterval = "week"
	IntervalMonth TimeInterval = "month"
)

// DataStyle represents the type of market data (ticks or candles)
type DataStyle string

const (
	StyleTicks   DataStyle = "ticks"
	StyleCandles DataStyle = "candles"
)

// HistoricalDataRequest represents parameters for historical data request
type HistoricalDataRequest struct {
	Symbol   string
	Interval TimeInterval // Time interval (hour, day, week, month)
	Style    DataStyle    // "ticks" or "candles"
	Count    int          // Number of ticks/candles to return
}

// HistoricalDataPoint represents a single historical data point
type HistoricalDataPoint struct {
	Timestamp int64
	Price     float64
	High      float64 // Only available for candles
	Low       float64 // Only available for candles
	Open      float64 // Only available for candles
	Close     float64 // Only available for candles
}
