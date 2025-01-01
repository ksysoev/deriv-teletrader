package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Basic command handlers
func (b *Bot) handleStart(ctx context.Context, msg *Message) (*Response, error) {
	text := `👋 Welcome to Deriv Trading Bot!

Use /help to see available commands.`
	return &Response{
		Text:             text,
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handleHelp(ctx context.Context, msg *Message) (*Response, error) {
	text := `Available commands:

/symbols - List available trading symbols
/balance - Show account balance
/price <symbol> - Get current price for a symbol
/buy <symbol> <amount> - Place a buy order
/sell <symbol> <amount> - Place a sell order
/position - Show current positions`
	return &Response{
		Text:             text,
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handleSymbols(ctx context.Context, msg *Message) (*Response, error) {
	symbols := strings.Join(b.symbols, "\n")
	text := fmt.Sprintf("Available symbols:\n\n%s", symbols)
	return &Response{
		Text:             text,
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handleBalance(ctx context.Context, msg *Message) (*Response, error) {
	balance, err := b.derivClient.GetBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return &Response{
		Text:             fmt.Sprintf("💰 Balance: %.2f %s", balance.Amount, balance.Currency),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handlePrice(ctx context.Context, msg *Message) (*Response, error) {
	if len(msg.Args) < 1 {
		return &Response{
			Text:             "❌ Please provide a symbol. Example: /price R_50",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	symbol := msg.Args[0]
	price, err := b.derivClient.GetPrice(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return &Response{
		Text:             fmt.Sprintf("💹 %s price: %.2f", symbol, price),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handleBuy(ctx context.Context, msg *Message) (*Response, error) {
	if len(msg.Args) < 2 {
		return &Response{
			Text:             "❌ Please provide symbol and amount. Example: /buy R_50 10.50",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	symbol := msg.Args[0]
	amount, err := strconv.ParseFloat(msg.Args[1], 64)
	if err != nil {
		return &Response{
			Text:             "❌ Invalid amount format. Please provide a number.",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	if err := b.derivClient.PlaceTrade(ctx, symbol, amount, "CALL"); err != nil {
		return nil, fmt.Errorf("failed to place buy order: %w", err)
	}

	return &Response{
		Text:             fmt.Sprintf("✅ Buy order placed for %s: $%.2f", symbol, amount),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handleSell(ctx context.Context, msg *Message) (*Response, error) {
	if len(msg.Args) < 2 {
		return &Response{
			Text:             "❌ Please provide symbol and amount. Example: /sell R_50 10.50",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	symbol := msg.Args[0]
	amount, err := strconv.ParseFloat(msg.Args[1], 64)
	if err != nil {
		return &Response{
			Text:             "❌ Invalid amount format. Please provide a number.",
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	if err := b.derivClient.PlaceTrade(ctx, symbol, amount, "PUT"); err != nil {
		return nil, fmt.Errorf("failed to place sell order: %w", err)
	}

	return &Response{
		Text:             fmt.Sprintf("✅ Sell order placed for %s: $%.2f", symbol, amount),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}

func (b *Bot) handlePosition(ctx context.Context, msg *Message) (*Response, error) {
	position, err := b.derivClient.GetPosition(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	return &Response{
		Text:             fmt.Sprintf("📊 Current positions:\n\n%s", position),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
	}, nil
}
