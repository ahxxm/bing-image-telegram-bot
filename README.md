# Bing Image Telegram Bot

[Channel](https://t.me/bing_image_global)

## Deploy Your Own

before `docker-compose up -d`, modify `docker-compose.yml`:

- `BOT_TOKEN` from `@BotFather`
- create a channel, add bot to channel, send message for next step
- `CHAT_ID` from `https://api.telegram.org/bot{BOT_TOKEN}/getUpdates`
- (optional) delete previous message from channel

## Image from Custom Markets

Current markets are from [here](https://www.microsoft.com/en-in/locale.aspx), all hardcoded in `main.go`. (credits to [Amar Palsapure](https://stackoverflow.com/questions/10639914/is-there-a-way-to-get-bings-photo-of-the-day#comment58369141_18096210))
