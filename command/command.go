package command

import (
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/message"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HelpInfo = "Справка по коммандам:\n/sh_start - Начать игру\n/sh_stop - Закончить игру и вывести результаты"
	Unknown  = "Неизвестная команда"
)

func HandleCommand(bot *tg.BotAPI, msg *tg.Message) {

	command := msg.Text[1:] // Get rid of prefix "/"
	
	if strings.HasPrefix(command, "sh_") {
		game.RunGameCommand(bot, msg.Chat, command)
		return
	}

	//Filter non-game commands e.g /help
	switch command {
	case "help":
		message.SendMessage(bot, msg.Chat, HelpInfo)
	default:
		message.SendMessage(bot, msg.Chat, Unknown)
	}
}
