package deriv

import (
	"context"
	"fmt"
	"strconv"

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
	var tickReq struct {
		Ticks     string `json:"ticks"`
		Subscribe int    `json:"subscribe"`
		ReqID     int    `json:"req_id"`
	}
	tickReq.Ticks = symbol
	tickReq.Subscribe = 0
	tickReq.ReqID = 2

	var tickResp struct {
		Tick struct {
			Quote float64 `json:"quote"`
		} `json:"tick"`
	}

	if err := c.api.SendRequest(ctx, 2, tickReq, &tickResp); err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	return tickResp.Tick.Quote, nil
}

// PlaceTrade places a trade order
func (c *Client) PlaceTrade(ctx context.Context, symbol string, amount float64, direction string) error {
	// First create a proposal
	var proposalReq struct {
		Proposal     int    `json:"proposal"`
		ContractType string `json:"contract_type"`
		Symbol       string `json:"symbol"`
		Duration     int    `json:"duration"`
		DurationUnit string `json:"duration_unit"`
		Basis        string `json:"basis"`
		Amount       string `json:"amount"`
		Currency     string `json:"currency"`
		ReqID        int    `json:"req_id"`
	}

	proposalReq.Proposal = 1
	proposalReq.ContractType = direction
	proposalReq.Symbol = symbol
	proposalReq.Duration = 60
	proposalReq.DurationUnit = "s"
	proposalReq.Basis = "stake"
	proposalReq.Amount = strconv.FormatFloat(amount, 'f', 2, 64)
	proposalReq.Currency = "USD"
	proposalReq.ReqID = 3

	var proposalResp struct {
		Proposal struct {
			ID string `json:"id"`
		} `json:"proposal"`
	}

	if err := c.api.SendRequest(ctx, 3, proposalReq, &proposalResp); err != nil {
		return fmt.Errorf("failed to create proposal: %w", err)
	}

	// Then buy the contract
	var buyReq struct {
		Buy   string  `json:"buy"`
		Price float64 `json:"price"`
		ReqID int     `json:"req_id"`
	}

	buyReq.Buy = proposalResp.Proposal.ID
	buyReq.Price = amount
	buyReq.ReqID = 4

	var buyResp struct {
		BuyContract struct {
			ContractID int `json:"contract_id"`
		} `json:"buy"`
	}

	if err := c.api.SendRequest(ctx, 4, buyReq, &buyResp); err != nil {
		return fmt.Errorf("failed to buy contract: %w", err)
	}

	return nil
}

// GetPosition retrieves current trading position
func (c *Client) GetPosition(ctx context.Context) (string, error) {
	var req struct {
		ProposalOpenContract int `json:"proposal_open_contract"`
		Subscribe            int `json:"subscribe"`
		ReqID                int `json:"req_id"`
	}
	req.ProposalOpenContract = 1
	req.Subscribe = 0
	req.ReqID = 5

	var resp struct {
		ProposalOpenContract []struct {
			ContractID   string  `json:"contract_id"`
			Symbol       string  `json:"symbol"`
			ContractType string  `json:"contract_type"`
			EntrySpot    float64 `json:"entry_spot"`
			CurrentSpot  float64 `json:"current_spot"`
			Profit       float64 `json:"profit"`
		} `json:"proposal_open_contract"`
	}

	if err := c.api.SendRequest(ctx, 5, req, &resp); err != nil {
		return "", fmt.Errorf("failed to get position: %w", err)
	}

	if len(resp.ProposalOpenContract) == 0 {
		return "No open positions", nil
	}

	var result string
	for _, contract := range resp.ProposalOpenContract {
		result += fmt.Sprintf("Contract ID: %s\nSymbol: %s\nType: %s\nEntry Spot: %.2f\nCurrent Spot: %.2f\nProfit: %.2f\n\n",
			contract.ContractID,
			contract.Symbol,
			contract.ContractType,
			contract.EntrySpot,
			contract.CurrentSpot,
			contract.Profit)
	}

	return result, nil
}
