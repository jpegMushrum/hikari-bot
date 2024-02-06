package game

import (
	"bakalover/hikari-bot/message"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const GreetingsString = "Начинается раунд"
const EndingString = "Результаты раунда:"

func IsRunning() bool {
	return GetCurrentGameState() == Running
}

func RunGameCommand(bot *tg.BotAPI, chat *tg.Chat, command string) {
	if ok, state := TryChangeState(command); ok {
		switch state {
		case Init:
			message.SendMessage(bot, chat, GreetingsString)
			// Init DB, send first word etc.
		case Running:
			message.SendMessage(bot, chat, EndingString)
			// Send statistics, prize places and reset DB
		}
	} else {
		switch state {
		case Init:
			message.SendMessage(bot, chat, "Игра не запущена!")
		case Running:
			message.SendMessage(bot, chat, "Игра уже запущена!")
		}
	}
}
