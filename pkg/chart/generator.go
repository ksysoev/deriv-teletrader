package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kirill/deriv-teletrader/pkg/types"
	"github.com/wcharczuk/go-chart/v2"
)

// GeneratePriceChart creates a price chart for the given historical data
func GeneratePriceChart(data []types.HistoricalDataPoint, symbol string) (string, error) {
	// Create temporary directory if it doesn't exist
	tmpDir := filepath.Join(os.TempDir(), "deriv-teletrader")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Prepare data for the chart
	var xValues []time.Time
	var yValues []float64

	for _, point := range data {
		xValues = append(xValues, time.Unix(point.Timestamp, 0))
		if point.Close != 0 {
			yValues = append(yValues, point.Close)
		} else {
			yValues = append(yValues, point.Price)
		}
	}

	// Create time series
	series := chart.TimeSeries{
		Name: symbol,
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
			StrokeWidth: 2,
		},
	}

	// Add data points
	for i := range xValues {
		series.XValues = append(series.XValues, xValues[i])
		series.YValues = append(series.YValues, yValues[i])
	}

	// Create chart with styling
	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:    20,
				Left:   20,
				Right:  20,
				Bottom: 20,
			},
		},
		XAxis: chart.XAxis{
			Name:           "Time",
			TickPosition:   chart.TickPositionBetweenTicks,
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04"),
			Style: chart.Style{
				StrokeWidth: 1,
				FontSize:    10,
			},
		},
		YAxis: chart.YAxis{
			Name: "Price",
			Style: chart.Style{
				StrokeWidth: 1,
				FontSize:    10,
			},
		},
		Series: []chart.Series{series},
	}

	// Add title
	graph.Title = fmt.Sprintf("%s Price Chart", symbol)

	// Create output file
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%d.png", symbol, time.Now().Unix()))
	f, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Render chart
	if err := graph.Render(chart.PNG, f); err != nil {
		return "", fmt.Errorf("failed to render chart: %w", err)
	}

	return outputPath, nil
}
