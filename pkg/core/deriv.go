package core

import (
	"context"
)

// DerivClient defines the interface for Deriv API operations
type DerivClient interface {
	Connect(ctx context.Context) error
	Close() error
	GetBalance(ctx context.Context) (float64, error)
	GetPrice(ctx context.Context, symbol string) (float64, error)
	PlaceTrade(ctx context.Context, symbol string, amount float64, direction string) error
	GetPosition(ctx context.Context) (string, error)
}
