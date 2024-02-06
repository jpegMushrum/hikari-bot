package message

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func SendMessage(bot *tg.BotAPI, chat *tg.Chat, msg string) {
	bot.Send(tg.NewMessage(chat.ID, msg))
}
