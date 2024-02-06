package main

import (
	"bakalover/hikari-bot/command"
	"bakalover/hikari-bot/game"
	"log"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const LogForm = "User: %s, Message: %s"

func LogMsg(msg *tg.Message) {
	log.Printf(LogForm, msg.From.UserName, msg.Text)
}
func main() {
	bot, err := tg.NewBotAPI(os.Getenv("HIKARI_BOT_TOKEN"))

	if err != nil {
		log.Panic("Error creating bot api!")
	}

	// bot.Debug = true

	uCfg := tg.NewUpdate(0) // No timeout (or maybe specify later)

	upds := bot.GetUpdatesChan(uCfg)

	for upd := range upds {
		if msg := upd.Message; msg != nil {

			LogMsg(msg)

			if msg.IsCommand() {
				command.HandleCommand(bot, msg)
			} else {
				if game.IsRunning() {
					// Check if user is new and add him
					// Else go and check word for chaining etc.
				}
			}
		}
	}

}
