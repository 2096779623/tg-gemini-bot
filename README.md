# tg-gemini-bot

A simply gemini chatbot on telegram.

# Installation

## git

```bash
git clone https://github.com/2096779623/tg-gemini-bot
cd tg-gemini-bot
go mod tidy
go build -o tg-gemini-bot
chmod +x tg-gemini-bot
TELEBOT_TOKEN=your_telegram_bot_token GEMINI_API_KEY=your_gemini_api_key ./tg-gemini-bot
# or `TELEBOT_TOKEN=your_telegram_bot_token GEMINI_API_KEY=your_gemini_api_key go run main.go` directly
```
## Docker (recommend)

```bash
docker pull ghcr.io/2096779623/tg-gemini-bot:latest
docker run --restart always --env TELEBOT_TOKEN=your_telegram_bot_token --env GEMINI_API_KEY=your_gemini_api_key --name tg-gemini-bot tg-gemini-bot:latest
```

# usage
```bash
Usage of tg-gemini-bot:
  -proxy string
        SOCKS5 proxy address
  -version bool
        show version
```
