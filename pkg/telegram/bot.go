package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirill/deriv-teletrader/pkg/core"
)

// MessageProcessor defines the interface for processing chat messages
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg *core.Message) (*core.Response, error)
}

// Config holds configuration specific to the Telegram bot
type Config struct {
	Token            string   `mapstructure:"token"`
	AllowedUsernames []string `mapstructure:"allowed_usernames"`
	Debug            bool     `mapstructure:"debug"`
}

type Bot struct {
	api       *tgbotapi.BotAPI
	processor MessageProcessor
}

// NewBot creates a new instance of the Telegram bot
func NewBot(cfg *Config, processor MessageProcessor) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = cfg.Debug

	bot := &Bot{
		api:       api,
		processor: processor,
	}

	return bot, nil
}

// Start begins polling for updates from Telegram
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if err := b.handleUpdate(ctx, update); err != nil {
				log.Printf("Error handling update: %v", err)
			}
		}
	}
}

// Stop gracefully shuts down the bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	var coreMsg *core.Message
	var chatID int64
	var messageID int

	// Handle callback queries from inline buttons
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		messageID = update.CallbackQuery.Message.MessageID

		coreMsg = &core.Message{
			ChatID:       chatID,
			MessageID:    messageID,
			Username:     update.CallbackQuery.From.UserName,
			CallbackData: update.CallbackQuery.Data,
		}

		// Answer callback query to remove loading state
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := b.api.Request(callback); err != nil {
			log.Printf("Failed to answer callback query: %v", err)
		}
	} else if update.Message != nil {
		// Handle regular messages
		msg := update.Message
		chatID = msg.Chat.ID
		messageID = msg.MessageID

		coreMsg = &core.Message{
			ChatID:    chatID,
			MessageID: messageID,
			Username:  msg.From.UserName,
		}

		// Handle commands
		if msg.IsCommand() {
			coreMsg.Command = msg.Command()
			coreMsg.Args = strings.Fields(msg.CommandArguments())
		} else if msg.Text != "" {
			// Handle regular messages by putting the text into Args
			coreMsg.Args = []string{msg.Text}
		}
	} else {
		// Skip other types of updates
		return nil
	}

	// Only show typing indicator for text messages, not callbacks
	if update.CallbackQuery == nil {
		// Send initial typing action immediately
		typingMsg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
		if _, err := b.api.Send(typingMsg); err != nil {
			log.Printf("Failed to send initial typing action: %v", err)
		}

		// Create a context with cancel for the typing goroutine
		typingCtx, cancelTyping := context.WithCancel(ctx)
		defer cancelTyping()

		// Start a goroutine to keep sending typing status periodically
		go func() {
			ticker := time.NewTicker(4 * time.Second) // Refresh typing status every 4 seconds
			defer ticker.Stop()

			for {
				select {
				case <-typingCtx.Done():
					return
				case <-ticker.C:
					typingMsg := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
					if _, err := b.api.Send(typingMsg); err != nil {
						log.Printf("Failed to send typing action: %v", err)
					}
				}
			}
		}()
	}

	// Process message using message processor
	response, err := b.processor.ProcessMessage(ctx, coreMsg)
	if err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}

	// Send response
	reply := tgbotapi.NewMessage(response.ChatID, response.Text)
	reply.ReplyToMessageID = response.ReplyToMessageID

	// Add inline keyboard if buttons are provided
	if len(response.Buttons) > 0 {
		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, row := range response.Buttons {
			var keyboardRow []tgbotapi.InlineKeyboardButton
			for _, btn := range row {
				keyboardRow = append(keyboardRow, tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.CallbackData))
			}
			keyboard = append(keyboard, keyboardRow)
		}
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	if _, err := b.api.Send(reply); err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return nil
}
