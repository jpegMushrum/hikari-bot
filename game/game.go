package game

import (
	"bakalover/hikari-bot/message"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	GreetingsString     = "ゲームラウンドが始まります!"
	EndingString        = "ラウンド結果:"
	IsNotStartedError   = "ゲームはまだ始まっていません!"
	AlreadyRunningError = "ゲームはもう始まっています！"
)

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
			message.SendMessage(bot, chat, IsNotStartedError)
		case Running:
			message.SendMessage(bot, chat, AlreadyRunningError)
		}
	}
}
