package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kirill/deriv-teletrader/config"
	"github.com/kirill/deriv-teletrader/internal/deriv"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	cfg             *config.Config
	derivClient     *deriv.Client
	allowedUsers    map[string]struct{}
	updatesChan     tgbotapi.UpdatesChannel
	commandHandlers map[string]CommandHandler
}

type CommandHandler func(ctx context.Context, msg *tgbotapi.Message) error

// NewBot creates a new instance of the Telegram bot
func NewBot(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = cfg.Debug

	// Initialize Deriv client
	derivClient, err := deriv.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Deriv client: %w", err)
	}

	// Create allowed users map for faster lookup
	allowedUsers := make(map[string]struct{})
	for _, username := range cfg.AllowedUsernames {
		allowedUsers[strings.ToLower(username)] = struct{}{}
	}

	bot := &Bot{
		api:          api,
		cfg:          cfg,
		derivClient:  derivClient,
		allowedUsers: allowedUsers,
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

// Start begins polling for updates from Telegram
func (b *Bot) Start(ctx context.Context) error {
	// Connect to Deriv API
	if err := b.derivClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to Deriv API: %w", err)
	}
	defer b.derivClient.Close()

	log.Printf("Connected to Deriv API")

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
	if err := b.derivClient.Close(); err != nil {
		log.Printf("Error closing Deriv client: %v", err)
	}
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	msg := update.Message

	// Check if user is allowed
	if !b.isUserAllowed(msg.From.UserName) {
		log.Printf("Unauthorized access attempt from @%s", msg.From.UserName)
		return b.reply(msg, "⚠️ You are not authorized to use this bot.")
	}

	// Handle commands
	if msg.IsCommand() {
		command := msg.Command()
		handler, exists := b.commandHandlers[command]
		if !exists {
			return b.reply(msg, "❌ Unknown command. Type /help for available commands.")
		}

		if err := handler(ctx, msg); err != nil {
			log.Printf("Error handling command %s: %v", command, err)
			return b.reply(msg, fmt.Sprintf("❌ Error executing command: %v", err))
		}
		return nil
	}

	// Handle non-command messages if needed
	return nil
}

// isUserAllowed checks if a user is allowed to use the bot
func (b *Bot) isUserAllowed(username string) bool {
	if username == "" {
		return false
	}
	_, allowed := b.allowedUsers[strings.ToLower(username)]
	return allowed
}

// reply sends a text message reply to the user
func (b *Bot) reply(msg *tgbotapi.Message, text string) error {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID
	_, err := b.api.Send(reply)
	return err
}

// Basic command handlers
func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message) error {
	text := `👋 Welcome to Deriv Trading Bot!

Use /help to see available commands.`
	return b.reply(msg, text)
}

func (b *Bot) handleHelp(ctx context.Context, msg *tgbotapi.Message) error {
	text := `Available commands:

/symbols - List available trading symbols
/balance - Show account balance
/price <symbol> - Get current price for a symbol
/buy <symbol> <amount> - Place a buy order
/sell <symbol> <amount> - Place a sell order
/position - Show current positions`
	return b.reply(msg, text)
}

func (b *Bot) handleSymbols(ctx context.Context, msg *tgbotapi.Message) error {
	symbols := strings.Join(b.cfg.DefaultSymbols, "\n")
	text := fmt.Sprintf("Available symbols:\n\n%s", symbols)
	return b.reply(msg, text)
}

func (b *Bot) handleBalance(ctx context.Context, msg *tgbotapi.Message) error {
	balance, err := b.derivClient.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	return b.reply(msg, fmt.Sprintf("💰 Balance: $%.2f", balance))
}

func (b *Bot) handlePrice(ctx context.Context, msg *tgbotapi.Message) error {
	args := strings.Fields(msg.CommandArguments())
	if len(args) < 1 {
		return b.reply(msg, "❌ Please provide a symbol. Example: /price R_50")
	}

	symbol := args[0]
	price, err := b.derivClient.GetPrice(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}

	return b.reply(msg, fmt.Sprintf("💹 %s price: %.2f", symbol, price))
}

func (b *Bot) handleBuy(ctx context.Context, msg *tgbotapi.Message) error {
	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		return b.reply(msg, "❌ Please provide symbol and amount. Example: /buy R_50 10.50")
	}

	symbol := args[0]
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return b.reply(msg, "❌ Invalid amount format. Please provide a number.")
	}

	if err := b.derivClient.PlaceTrade(ctx, symbol, amount, "CALL"); err != nil {
		return fmt.Errorf("failed to place buy order: %w", err)
	}

	return b.reply(msg, fmt.Sprintf("✅ Buy order placed for %s: $%.2f", symbol, amount))
}

func (b *Bot) handleSell(ctx context.Context, msg *tgbotapi.Message) error {
	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		return b.reply(msg, "❌ Please provide symbol and amount. Example: /sell R_50 10.50")
	}

	symbol := args[0]
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return b.reply(msg, "❌ Invalid amount format. Please provide a number.")
	}

	if err := b.derivClient.PlaceTrade(ctx, symbol, amount, "PUT"); err != nil {
		return fmt.Errorf("failed to place sell order: %w", err)
	}

	return b.reply(msg, fmt.Sprintf("✅ Sell order placed for %s: $%.2f", symbol, amount))
}

func (b *Bot) handlePosition(ctx context.Context, msg *tgbotapi.Message) error {
	position, err := b.derivClient.GetPosition(ctx)
	if err != nil {
		return fmt.Errorf("failed to get position: %w", err)
	}
	return b.reply(msg, fmt.Sprintf("📊 Current positions:\n\n%s", position))
}
