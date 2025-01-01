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
/buy <symbol> <amount> - Place a trade (Up/Down)
/position - Show current positions

Example:
1. /buy R_50 10.50
2. Select Up ⬆️ or Down ⬇️`
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
	// If there's callback data, handle the direction selection
	if msg.CallbackData != "" {
		data := ParseCallbackData(msg.CallbackData)
		if data["action"] != "trade" {
			return nil, fmt.Errorf("invalid callback action: %s", data["action"])
		}

		symbol := data["symbol"]
		amount, err := strconv.ParseFloat(data["amount"], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount in callback: %w", err)
		}

		direction := "CALL"
		if strings.HasSuffix(msg.CallbackData, ":down") {
			direction = "PUT"
		}

		if err := b.derivClient.PlaceTrade(ctx, symbol, amount, direction); err != nil {
			return nil, fmt.Errorf("failed to place trade: %w", err)
		}

		directionEmoji := "⬆️"
		if direction == "PUT" {
			directionEmoji = "⬇️"
		}

		return &Response{
			Text:             fmt.Sprintf("✅ %s Trade placed for %s: $%.2f", directionEmoji, symbol, amount),
			ReplyToMessageID: msg.MessageID,
			ChatID:           msg.ChatID,
		}, nil
	}

	// Initial /buy command handling
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

	// Create callback data with trade details
	callbackBase := fmt.Sprintf("trade:%s:%.2f", symbol, amount)

	// Create Up/Down buttons
	buttons := [][]Button{
		{
			{Text: "Up ⬆️", CallbackData: callbackBase + ":up"},
			{Text: "Down ⬇️", CallbackData: callbackBase + ":down"},
		},
	}

	return &Response{
		Text:             fmt.Sprintf("🎯 Place a trade for %s: $%.2f\nSelect direction:", symbol, amount),
		ReplyToMessageID: msg.MessageID,
		ChatID:           msg.ChatID,
		Buttons:          buttons,
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
