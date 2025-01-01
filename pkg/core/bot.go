package core

import (
	"context"
	"fmt"
	"strings"
)

// BalanceInfo contains balance amount and currency
type BalanceInfo struct {
	Amount   float64
	Currency string
}

// DerivClient defines the interface for Deriv API operations
type DerivClient interface {
	MarketDataProvider
	GetBalance(ctx context.Context) (*BalanceInfo, error)
	PlaceTrade(ctx context.Context, symbol string, amount float64, direction string) error
	GetPosition(ctx context.Context) (string, error)
}

// Message represents a chat message with parsed command and arguments
type Message struct {
	Command   string
	Args      []string
	ChatID    int64
	MessageID int
	Username  string
}

// Response represents a response to a chat message
type Response struct {
	Text             string
	ReplyToMessageID int
	ChatID           int64
}

// Bot handles the business logic for processing chat messages
type Bot struct {
	derivClient     DerivClient
	llmClient       LLMClient
	allowedUsers    map[string]struct{}
	commandHandlers map[string]CommandHandler
	symbols         []string
}

type CommandHandler func(ctx context.Context, msg *Message) (*Response, error)

// NewBot creates a new instance of the bot
func NewBot(derivClient DerivClient, llmClient LLMClient, allowedUsers []string, symbols []string) (*Bot, error) {

	// Create allowed users map for faster lookup
	allowedUsersMap := make(map[string]struct{})
	for _, username := range allowedUsers {
		allowedUsersMap[username] = struct{}{}
	}

	bot := &Bot{
		derivClient:  derivClient,
		llmClient:    llmClient,
		allowedUsers: allowedUsersMap,
		symbols:      symbols,
	}

	// Initialize command handlers
	bot.commandHandlers = map[string]CommandHandler{
		"start":    bot.handleStart,
		"help":     bot.handleHelp,
		"symbols":  bot.handleSymbols,
		"balance":  bot.handleBalance,
		"price":    bot.handlePrice,
		"buy":      bot.handleBuy,
		"sell":     bot.handleSell,
		"position": bot.handlePosition,
	}

	return bot, nil
}

// ProcessMessage processes an incoming message and returns a response
func (b *Bot) ProcessMessage(ctx context.Context, msg *Message) (*Response, error) {
	// Check if user is allowed
	if !b.isUserAllowed(msg.Username) {
		return &Response{
			Text:             "⚠️ You are not authorized to use this bot.",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	// If it's not a command (doesn't start with /), process as free-form text
	if msg.Command == "" {
		// Join all args as they represent the full message text
		text := strings.Join(msg.Args, " ")
		if text == "" {
			return &Response{
				Text:             "❌ Please provide some text for me to process.",
				ReplyToMessageID: msg.MessageID,
				ChatID:           msg.ChatID,
			}, nil
		}

		// Process text with LLM using market data functions
		response, err := b.llmClient.ProcessWithFunctions(ctx, text, b.derivClient, MarketDataFunctions)
		if err != nil {
			return nil, fmt.Errorf("failed to process text: %w", err)
		}

		return &Response{
			Text:             response,
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	// Handle command
	handler, exists := b.commandHandlers[msg.Command]
	if !exists {
		return &Response{
			Text:             "❌ Unknown command. Type /help for available commands.",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	return handler(ctx, msg)
}

// isUserAllowed checks if a user is allowed to use the bot
func (b *Bot) isUserAllowed(username string) bool {
	if username == "" {
		return false
	}
	_, allowed := b.allowedUsers[username]
	return allowed
}
