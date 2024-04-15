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

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	HelpInfo = `
	Справка по коммандам:
	/help - Справка по командам
	/rules - Правила игры
	/sh_start - Начать игру
	/sh_stop - Закончить игру и вывести результаты`

	Unknown         = "Неизвестная команда"
	ShiritoryPrefix = "sh_"
	Rules           = `
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
)

func HandleCommand(dbConn *gorm.DB, bot *tg.BotAPI, msg *tg.Message) {
	command := msg.Command()

	if strings.HasPrefix(command, ShiritoryPrefix) {
		game.RunGameCommand(util.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg})
		return
	}

	//Filter non-game commands e.g /help
	switch command {
	case "help":
		util.Reply(util.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg}, HelpInfo)
	case "rules":
		util.Reply(util.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg}, Rules)

	default:
		util.Reply(util.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg}, Unknown)
	}
}

func connectToPostgres(dsn string) (*gorm.DB, error) {
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

	bot, err := tg.NewBotAPI(os.Getenv("HIKARI_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Couldn't initialize bot api!\n%v", err)
	}

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=5432 sslmode=disable",
		os.Getenv("PG_HOST"), os.Getenv("PG_LOGIN"), os.Getenv("PG_PASS"), os.Getenv("PG_DB"))
	dbConn, err := connectToPostgres(dsn)
	if err != nil {
		log.Fatalf("Couldn't establish connection to PostgreSQL!\n%v", err)
	}

	// bot.Debug = true

	uCfg := tg.NewUpdate(0) // No timeout (or maybe specify later)
	uCfg.Timeout = 60
	uCfg.AllowedUpdates = []string{"message"}

	// Strand | MPSC
	upds := bot.GetUpdatesChan(uCfg)

	dict := &jisho.JishoDict{}

	for upd := range upds {
		if msg := upd.Message; msg != nil {
			log.Printf("User: %v, Message: %v", msg.From.UserName, msg.Text)
			if msg.IsCommand() {
				HandleCommand(dbConn, bot, msg)
			} else {
				log.Println(game.Chat())
				log.Println(msg.Chat.ID)
				if game.Chat() == msg.Chat.ID && game.IsRunning() {
					game.HandleNextWord(util.MsgContext{DbConn: dbConn, Bot: bot, Msg: msg}, dict)
				}
			}
		}
	}
}
