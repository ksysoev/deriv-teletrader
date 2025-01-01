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
	result += fmt.Sprintf("Contract ID: %s\nType: %s\nEntry Spot: %.2f\nCurrent Spot: %.2f\nProfit: %.2f\n\n",
		resp.ProposalOpenContract.ContractId,
		resp.ProposalOpenContract.ContractType,
		*resp.ProposalOpenContract.EntrySpot,
		*resp.ProposalOpenContract.CurrentSpot,
		*resp.ProposalOpenContract.Profit)

	return result, nil
}
