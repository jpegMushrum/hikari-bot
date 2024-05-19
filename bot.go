package main

import (
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func connectToDatabase(dsn string) (*gorm.DB, error) {
	const maxRetries = 3
	const delayBetweenRetries = time.Second

	var dbConn *gorm.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			return dbConn, nil
		}
		log.Printf("Couldn't establish connection to PostgreSQL (attempt %d), retrying in %v...\n%v", i+1, delayBetweenRetries, err)
		time.Sleep(delayBetweenRetries)
	}
	return nil, err
}

func main() {

	bot, err := tele.NewBot(tele.Settings{
		Token:       os.Getenv("HIKARI_BOT_TOKEN"),
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: true, // Bottleneck
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

	dbConn, err := connectToDatabase(dsn)
	if err != nil {
		log.Fatalf("Couldn't establish connection to Database!\n%v", err)
	}

	dict := &jisho.JishoDict{}

	bot.Handle("/help", func(c tele.Context) error {
		util.Reply(c, HelpInfo)
		return nil
	})

	bot.Handle("/rules", func(c tele.Context) error {
		util.Reply(c, Rules)
		return nil
	})

	bot.Handle("/start_game", func(c tele.Context) error {
		game.HandleCommand(util.GameContext{DbConn: dbConn, TeleCtx: c})
		return nil
	})

	bot.Handle("/stop_game", func(c tele.Context) error {
		game.HandleCommand(util.GameContext{DbConn: dbConn, TeleCtx: c})
		return nil
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		if strings.HasPrefix(c.Text(), "/") { // Filter unused commands
			util.Reply(c, UnknownCommand)
			return nil
		}
		if game.Thread() == c.Message().ThreadID {
			game.HandleNextWord(util.GameContext{DbConn: dbConn, TeleCtx: c}, dict)
		}
		return nil
	})

	bot.Start()
}
