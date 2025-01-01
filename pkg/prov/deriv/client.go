package deriv

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kirill/deriv-teletrader/pkg/core"
	deriv "github.com/ksysoev/deriv-api"
	"github.com/ksysoev/deriv-api/schema"
)

// Config holds Deriv-specific configuration
type Config struct {
	AppID    string   `mapstructure:"app_id"`
	APIToken string   `mapstructure:"api_token"`
	Endpoint string   `mapstructure:"endpoint"`
	Symbols  []string `mapstructure:"symbols"`
}

type Client struct {
	api *deriv.Client
	cfg *Config
}

// NewClient creates a new Deriv API client
func NewClient(cfg *Config) (*Client, error) {
	appID, err := strconv.Atoi(cfg.AppID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	api, err := deriv.NewDerivAPI(
		cfg.Endpoint,
		appID,
		"en",
		"https://deriv-teletrader",
		deriv.Debug,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Client{
		api: api,
		cfg: cfg,
	}, nil
}

// Connect establishes connection to Deriv API and authorizes the session
func (c *Client) Connect(ctx context.Context) error {
	if err := c.api.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Authorize the connection
	reqAuth := schema.Authorize{Authorize: c.cfg.APIToken}
	if _, err := c.api.Authorize(ctx, reqAuth); err != nil {
		c.api.Disconnect()
		return fmt.Errorf("failed to authorize: %w", err)
	}

	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	c.api.Disconnect()
	return nil
}

// GetAvailableSymbols returns a list of available trading symbols
func (c *Client) GetAvailableSymbols(ctx context.Context) ([]string, error) {
	return c.cfg.Symbols, nil
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context) (*core.BalanceInfo, error) {
	req := schema.Balance{Balance: 1}

	resp, err := c.api.Balance(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &core.BalanceInfo{
		Amount:   resp.Balance.Balance,
		Currency: resp.Balance.Currency,
	}, nil
}

// GetPrice retrieves current price for a symbol
func (c *Client) GetPrice(ctx context.Context, symbol string) (float64, error) {
	req := schema.Ticks{
		Ticks: symbol,
	}

	resp, err := c.api.Ticks(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	if resp.Tick.Quote == nil {
		return 0, fmt.Errorf("no quote available")
	}
	return *resp.Tick.Quote, nil
}

// PlaceTrade places a trade order
func (c *Client) PlaceTrade(ctx context.Context, symbol string, amount float64, direction string) error {
	// Create a proposal
	duration := 5
	basis := schema.ProposalBasisStake

	// Convert direction string to ProposalContractType
	var contractType schema.ProposalContractType
	if direction == "CALL" {
		contractType = schema.ProposalContractTypeCALL
	} else {
		contractType = schema.ProposalContractTypePUT
	}

	req := schema.Proposal{
		Proposal:     1,
		Amount:       &amount,
		Basis:        &basis,
		ContractType: contractType,
		Currency:     "USD",
		Duration:     &duration,
		DurationUnit: "t",
		Symbol:       symbol,
	}

	resp, err := c.api.Proposal(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create proposal: %w", err)
	}

	// Buy the contract
	buyReq := schema.Buy{
		Buy:   resp.Proposal.Id,
		Price: amount,
	}

	_, err = c.api.Buy(ctx, buyReq)
	if err != nil {
		return fmt.Errorf("failed to buy contract: %w", err)
	}

	return nil
}

// convertDataStyle converts core.DataStyle to schema.TicksHistoryStyle
func convertDataStyle(style core.DataStyle) schema.TicksHistoryStyle {
	switch style {
	case core.StyleCandles:
		return schema.TicksHistoryStyleCandles
	default:
		return schema.TicksHistoryStyleTicks
	}
}

// GetHistoricalData retrieves historical market data for a given symbol and time period
func (c *Client) GetHistoricalData(ctx context.Context, req core.HistoricalDataRequest) ([]core.HistoricalDataPoint, error) {
	// Calculate start time based on interval
	now := time.Now().Unix()
	startTime := int(now)

	switch req.Interval {
	case core.IntervalHour:
		startTime -= 3600 // 1 hour ago
	case core.IntervalDay:
		startTime -= 86400 // 24 hours ago
	case core.IntervalWeek:
		startTime -= 604800 // 7 days ago
	case core.IntervalMonth:
		startTime -= 2592000 // 30 days ago
	default:
		return nil, fmt.Errorf("invalid interval: %s", req.Interval)
	}

	granularity := schema.TicksHistoryGranularity(60) // 1 minute candles by default
	style := convertDataStyle(req.Style)

	// Prepare the tick history request
	historyReq := schema.TicksHistory{
		TicksHistory: req.Symbol,
		End:          "latest",   // Always get data up to current time
		Start:        &startTime, // Pass pointer to integer
		Style:        style,
		Count:        req.Count,
		Granularity:  &granularity,
	}

	resp, err := c.api.TicksHistory(ctx, historyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	var result []core.HistoricalDataPoint

	if style == schema.TicksHistoryStyleTicks && resp.History != nil && resp.History.Prices != nil {
		for i, timestamp := range resp.History.Times {
			if i < len(resp.History.Prices) {
				result = append(result, core.HistoricalDataPoint{
					Timestamp: int64(timestamp),
					Price:     resp.History.Prices[i],
				})
			}
		}
	} else if style == schema.TicksHistoryStyleCandles && resp.Candles != nil {
		for _, candle := range resp.Candles {
			if candle.Epoch != nil && candle.Close != nil {
				point := core.HistoricalDataPoint{
					Timestamp: int64(*candle.Epoch),
					Price:     *candle.Close,
					Close:     *candle.Close,
				}

				if candle.Open != nil {
					point.Open = *candle.Open
				}
				if candle.High != nil {
					point.High = *candle.High
				}
				if candle.Low != nil {
					point.Low = *candle.Low
				}

				result = append(result, point)
			}
		}
	}

	return result, nil
}

// GetPosition retrieves current trading position
func (c *Client) GetPosition(ctx context.Context) (string, error) {
	req := schema.ProposalOpenContract{
		ProposalOpenContract: 1,
	}

	resp, err := c.api.ProposalOpenContract(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get position: %w", err)
	}

	if resp.ProposalOpenContract == nil {
		return "No open positions", nil
	}

	var result string
	result += fmt.Sprintf("Contract ID: %d\nType: %s\nEntry Spot: %.2f\nCurrent Spot: %.2f\nProfit: %.2f\n\n",
		*resp.ProposalOpenContract.ContractId,
		*resp.ProposalOpenContract.ContractType,
		*resp.ProposalOpenContract.EntrySpot,
		*resp.ProposalOpenContract.CurrentSpot,
		*resp.ProposalOpenContract.Profit)

	return result, nil
}
