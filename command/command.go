package command

import (
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/message"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const HelpInfo = "Справка по коммандам:\n/start - Начать игру\n/stop - Закончить игру и вывести результаты"

func HandleCommand(bot *tg.BotAPI, msg *tg.Message) {

	command := msg.Text[1:] // Get rid of prefix "/"

	//Filter non-game commands e.g /help
	switch command {
	case "help":
		message.SendMessage(bot, msg.Chat, HelpInfo)
	default:
		game.RunGameCommand(bot, msg.Chat, command)
	}
}
