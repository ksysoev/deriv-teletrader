package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirill/deriv-teletrader/pkg/core"
)

// Config holds configuration specific to the Telegram bot
type Config struct {
	Token            string   `mapstructure:"token"`
	AllowedUsernames []string `mapstructure:"allowed_usernames"`
	Debug            bool     `mapstructure:"debug"`
}

type Bot struct {
	api     *tgbotapi.BotAPI
	coreBot *core.Bot
}

// NewBot creates a new instance of the Telegram bot
func NewBot(cfg *Config, coreBot *core.Bot) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = cfg.Debug

	bot := &Bot{
		api:     api,
		coreBot: coreBot,
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
	}

	// Process message using core bot
	response, err := b.coreBot.ProcessMessage(ctx, coreMsg)
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
