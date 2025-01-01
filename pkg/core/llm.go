package core

import "context"

// LLMClient defines the interface for LLM operations
type LLMClient interface {
	ProcessText(ctx context.Context, input string) (string, error)
}
