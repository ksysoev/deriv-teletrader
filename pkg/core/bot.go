package core

import (
	"context"
	"fmt"

	"github.com/kirill/deriv-teletrader/pkg/deriv"
)

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
	derivClient     *deriv.Client
	allowedUsers    map[string]struct{}
	commandHandlers map[string]CommandHandler
	symbols         []string
}

type CommandHandler func(ctx context.Context, msg *Message) (*Response, error)

// NewBot creates a new instance of the bot
func NewBot(derivCfg *deriv.Config, allowedUsers []string, symbols []string) (*Bot, error) {
	// Initialize Deriv client
	derivClient, err := deriv.NewClient(derivCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Deriv client: %w", err)
	}

	// Create allowed users map for faster lookup
	allowedUsersMap := make(map[string]struct{})
	for _, username := range allowedUsers {
		allowedUsersMap[username] = struct{}{}
	}

	bot := &Bot{
		derivClient:  derivClient,
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

// Connect connects to the Deriv API
func (b *Bot) Connect(ctx context.Context) error {
	return b.derivClient.Connect(ctx)
}

// Close closes the connection to the Deriv API
func (b *Bot) Close() error {
	return b.derivClient.Close()
}
