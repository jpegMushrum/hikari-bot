package main

import (
	"bakalover/hikari-bot/game"
	"database/sql"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	HelpInfo        = "Справка по коммандам:\nhelp - Справка по командам\n/!start - Начать игру\n/!stop - Закончить игру и вывести результаты"
	Unknown         = "Неизвестная команда"
	ShiritoryPrefix = "!"
)

func HandleCommand(dbConn *sql.DB, bot *tg.BotAPI, msg *tg.Message) {
	command := msg.Command()

	if strings.HasPrefix(command, ShiritoryPrefix) {
		game.RunGameCommand(game.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg})
		return
	}

	//Filter non-game commands e.g /help
	switch command {
	case "help":
		bot.Send(tg.NewMessage(msg.Chat.ID, HelpInfo))

	default:
		bot.Send(tg.NewMessage(msg.Chat.ID, Unknown))
	}
}
