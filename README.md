# Deriv Teletrader

A Telegram bot for trading on Deriv platform using their API. The bot provides a simple interface for executing trades and monitoring positions through Telegram commands.

## Features

- Real-time trading on Deriv platform
- Account balance monitoring
- Price checking for trading symbols
- Position tracking
- Secure access with authorized users only

## Setup

1. Clone the repository:
```bash
git clone https://github.com/kirill/deriv-teletrader.git
cd deriv-teletrader
```

2. Create a configuration file by copying the example:
```bash
cp config.example.yaml config.yaml
```

3. Edit `config.yaml` and fill in your credentials:
- Telegram bot token (get it from [@BotFather](https://t.me/BotFather))
- List of authorized Telegram usernames
- Deriv API credentials (get them from [Deriv API](https://app.deriv.com/account/api-token))

4. Build the bot:
```bash
go build
```

5. Run the bot:
```bash
./deriv-teletrader start
```

## Configuration

The bot can be configured using a YAML file or environment variables. Here's an example configuration:

```yaml
# Telegram Bot Configuration
telegram:
  token: "your_telegram_bot_token"
  allowed_usernames:
    - "your_telegram_username"
  debug: false

# Deriv API Configuration
deriv:
  app_id: "your_deriv_app_id"
  api_token: "your_deriv_api_token"
  endpoint: "wss://ws.binaryws.com/websockets/v3"
  symbols:
    - "R_10"
    - "R_25"
    - "R_50"
    - "R_75"
    - "R_100"
```

Environment variables can be used with the prefix `TELETRADER_`, for example:
- `TELETRADER_TELEGRAM_TOKEN`
- `TELETRADER_TELEGRAM_ALLOWED_USERNAMES`
- `TELETRADER_DERIV_APP_ID`
- etc.

## Available Commands

- `/start` - Welcome message and bot introduction
- `/help` - Show available commands
- `/symbols` - List available trading symbols
- `/balance` - Show account balance
- `/price <symbol>` - Get current price for a symbol
- `/buy <symbol> <amount>` - Place a buy order
- `/sell <symbol> <amount>` - Place a sell order
- `/position` - Show current positions

## Examples

1. Check balance:
```
/balance
```

2. Get price for a symbol:
```
/price R_50
```

3. Place a buy order:
```
/buy R_50 10.50
```

4. Place a sell order:
```
/sell R_50 10.50
```

## Development

## Project Structure

```
.
├── pkg/           
│   ├── cmd/       # Command-line interface and configuration handling
│   ├── core/      # Core business logic and message processing
│   ├── prov/      # External service providers
│   │   └── deriv/ # Deriv API client implementation
│   └── telegram/  # Telegram bot implementation with its own config
└── config.yaml    # Configuration file
```

The project follows a modular structure:
- `pkg/cmd`: Contains CLI commands, configuration handling, and manages service lifecycles
- `pkg/core`: Implements core business logic and message processing in a stateless manner
- `pkg/prov`: Contains external service provider implementations
  - `pkg/prov/deriv`: Implements the Deriv API client
- `pkg/telegram`: Implements the Telegram bot with its own configuration structure

## Technologies

- [Cobra](https://github.com/spf13/cobra) for CLI commands
- [Viper](https://github.com/spf13/viper) for configuration
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) for Telegram integration
- [deriv-api](https://github.com/ksysoev/deriv-api) for Deriv API integration

## License

MIT License - see [LICENSE](LICENSE) for details.
