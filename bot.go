package main

import (
	"bakalover/hikari-bot/controller"
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/dict/jisho"
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	tele "gopkg.in/telebot.v3"
)

const (
	HelpInfo = `
	Справка по коммандам:
	/help - Справка по командам
	/rules - Правила игры
	/start_game - Начать игру
	/stop_game - Закончить игру и вывести результаты`

	Unknown = "Неизвестная команда"
	Rules   = `
	Правила:
	1. Два или более человек по очереди играют.

	2. Допускаются только существительные.

	3. Игрок, который выбирает слово, оканчивающееся на ん, 
	проигрывает игру, поскольку японское слово не начинается с 
	этого символа.
	
	4. Слова не могут повторяться.
	Пример: 
	-> сакура	(さくら)	
	-> радио    (ラジオ) 	
	-> онигири  (おにぎり)	
	-> рису 	(りす)		
	-> сумо 	(すもう)
	-> удон 	(うどん)
	Дополнительно: для удобства можно вводить слова как в форме кандзи так и в чистой кане`

	UnknownCommand = "Неизвестная комманда"
)

func main() {
	log.Println("Running hikari-bot v1.1.2")

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

	handlerComposit.AddHandler("/help", &controller.SimpleHandler{
		Inner: func(c util.GameContext) error {
			util.Reply(c.TeleCtx, HelpInfo)
			return nil
		},
	})

	handlerComposit.AddHandler("/rules", &controller.SimpleHandler{
		Inner: func(c util.GameContext) error {
			util.Reply(c.TeleCtx, Rules)
			return nil
		},
	})

	handlerComposit.AddHandler("/start_game", &controller.SimpleHandler{
		Inner: func(c util.GameContext) error {
			game.HandleCommand(c)
			return nil
		},
	})

	handlerComposit.AddHandler("/stop_game", &controller.SimpleHandler{
		Inner: func(c util.GameContext) error {
			game.HandleCommand(c)
			return nil
		},
	})

	handlerComposit.AddHandler(".", &controller.SimpleHandler{
		Inner: func(c util.GameContext) error {
			ctx := c.TeleCtx

			if strings.HasPrefix(ctx.Text(), "/") { // Filter unused commands
				util.Reply(ctx, UnknownCommand)
				return nil
			}

			if game.Thread() == ctx.Message().ThreadID {
				game.HandleNextWord(c)
			}

			return nil
		},
	})

	seaker := controller.NewOverseer(handlerComposit)

	bot.Handle(tele.OnText, func(c tele.Context) error {
		log.Printf("Handling message: %s\nfrom chat %v, thread %v, user %s", c.Text(), c.Chat().ID, c.Message().ThreadID, c.Sender().FirstName)

		ctk := util.GetCTK(c)
		seaker.SendMessage(util.GameContext{CTK: ctk, DbConn: dbConn, TeleCtx: c, Dicts: dicts})

		return nil
	})

	bot.Start()
}
