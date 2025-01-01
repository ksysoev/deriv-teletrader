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
			if update.Message == nil {
				continue
			}

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
	msg := update.Message

	// Convert Telegram message to core Message
	coreMsg := &core.Message{
		ChatID:    msg.Chat.ID,
		MessageID: msg.MessageID,
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

	// Send initial typing action immediately
	typingMsg := tgbotapi.NewChatAction(msg.Chat.ID, tgbotapi.ChatTyping)
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
				typingMsg := tgbotapi.NewChatAction(msg.Chat.ID, tgbotapi.ChatTyping)
				if _, err := b.api.Send(typingMsg); err != nil {
					log.Printf("Failed to send typing action: %v", err)
				}
			}
		}
	}()

	// Process message using message processor
	response, err := b.processor.ProcessMessage(ctx, coreMsg)
	if err != nil {
		return fmt.Errorf("failed to process message: %w", err)
	}

	// Send response
	reply := tgbotapi.NewMessage(response.ChatID, response.Text)
	reply.ReplyToMessageID = response.ReplyToMessageID
	if _, err := b.api.Send(reply); err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return nil
}
