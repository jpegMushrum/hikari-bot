package main

import (
	"bakalover/hikari-bot/controller"
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/dict/jisho"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	tele "gopkg.in/telebot.v3"
)

func main() {
	log.Println("Running hikari-bot v1.1.4")

	bot, err := tele.NewBot(tele.Settings{
		Token:       os.Getenv("HIKARI_BOT_TOKEN"),
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: true,
	})

	if err != nil {
		log.Fatalf("Couldn't initialize bot api!\n%v", err)
	}

	dsn := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=5432 sslmode=disable",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_LOGIN"),
		os.Getenv("PG_PASS"),
		os.Getenv("PG_DB"),
	)

	dbConn, err := dao.NewConnection(dsn)
	if err != nil {
		log.Fatalf("Couldn't establish connection to Database!\n%v", err)
	} else {
		log.Println("Database connection established!")
	}

	// Info sources
	dicts := []dict.Dictionary{
		jisho.NewJisho(), // DO NOT REORDER
		// Here goes JMDict and other
	}

	handlerComposit := controller.NewHandlerComposit()
	handlerComposit.AddHandler("/help", &controller.HelpHandler{})
	handlerComposit.AddHandler("/rules", &controller.RulesHandler{})
	handlerComposit.AddHandler("/start_game", &controller.StartGameHandler{})
	handlerComposit.AddHandler("/stop_game", &controller.StopGameHandler{})
	handlerComposit.AddHandler(".", &controller.NextWordGameHandler{})

	overseer := controller.NewOverseer(handlerComposit, dicts, dbConn)

	bot.Handle(tele.OnText, func(c tele.Context) error {
		log.Printf("Handling message: %s\nfrom chat %v, thread %v, user %s", c.Text(), c.Chat().ID, c.Message().ThreadID, c.Sender().FirstName)

		overseer.SendMessage(c)

		return nil
	})

	bot.Start()
}
