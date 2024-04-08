package main

import (
	"bakalover/hikari-bot/db"
	"bakalover/hikari-bot/dict/jisho"
	"bakalover/hikari-bot/game"
	"database/sql"
	"fmt"
	"log"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func main() {

	bot, err := tg.NewBotAPI(os.Getenv("HIKARI_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Couldn't initialize bot api!\n%v", err)
	}

	dbConn, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s",
			os.Getenv("PG_URL"), //localhost:35080/mytest
			os.Getenv("PG_LOGIN"),
			os.Getenv("PG_PASS"),
		),
	)

	if err != nil {
		log.Fatalf("Couldn't establish connection to PostgreSQL!\n%v", err)
	}

	// bot.Debug = true

	uCfg := tg.NewUpdate(0) // No timeout (or maybe specify later)

	// Strand | MPSC
	upds := bot.GetUpdatesChan(uCfg)

	dict := &jisho.JishoDict{}

	for upd := range upds {
		if msg := upd.Message; msg != nil {
			log.Printf("User: %v, Message: %v", msg.From.UserName, msg.Text)
			if msg.IsCommand() {
				HandleCommand(dbConn, bot, msg)
			} else {
				if game.Chat() == msg.Chat.ID && game.IsRunning() {
					game.HandleNextWord(game.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg}, dict)
				}
			}
		}
	}

	db.ExecuteScript(dbConn, "delete")
}
