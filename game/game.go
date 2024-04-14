package game

import (
	"bakalover/hikari-bot/db"
	"bakalover/hikari-bot/dict/jisho"
	"fmt"
	"log"
	"math/rand"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	GreetingsString     = "Раунд начинается!"
	EndingString        = "Результаты раунда:"
	IsNotStartedError   = "Игра ещё не началась!"
	AlreadyRunningError = "Игра уже запущена！"
)

const (
	word1 = "めいし"
	word2 = "りんご"
	word3 = "しお"
	word4 = "にんぎょう"
	word5 = "にっき"
	word6 = "しゅうまつ"
)

func Send(bot *tg.BotAPI, msg string) {
	bot.Send(tg.NewMessage(Chat(), msg))
}

func RandomizeStart(ctx MsgContext) {
	words := []string{word1, word2, word3, word4, word5, word6}
	initWord := words[rand.Intn(len(words))]
	db.AddWord(ctx.DbConn, initWord, "DUMMY_USER")
	Send(ctx.Bot, fmt.Sprintf("Первое слово: %s", initWord))
}

func AddPlayer(ctx MsgContext) {
	from := ctx.Msg.From
	db.AddPlayer(ctx.DbConn, from.UserName)
	Send(ctx.Bot, fmt.Sprintf("%s, добро пожаловать в игру!", from.FirstName))
}

func PlayerExists(ctx MsgContext) bool {
	return db.CheckPlayerExistence(ctx.DbConn, ctx.Msg.From.UserName)
}

func RunGameCommand(ctx MsgContext) {
	if ok, state := ExchangeState(ctx.Msg.Command()); ok {
		switch state {
		case Init:
			SetChat(ctx.Msg.Chat.ID)
			db.Init(ctx.DbConn)
			AddPlayer(ctx) // Player who pressed sh_start
			Send(ctx.Bot, GreetingsString)
			RandomizeStart(ctx)
		case Running:
			Send(ctx.Bot, EndingString)
			// FormAndSendStat(ctx)
			db.ShutDown(ctx.DbConn)
		}
	} else {
		switch state {
		case Init:
			Send(ctx.Bot, IsNotStartedError)
		case Running:
			Send(ctx.Bot, AlreadyRunningError)
		}
	}
}

func HandleNextWord(ctx MsgContext, dict *jisho.JishoDict) {
	if !PlayerExists(ctx) {
		AddPlayer(ctx)
	}

	maybeNextWord := ctx.Msg.Text

	if IsJapSuitable(maybeNextWord) {
		lastWord := db.GetLastWord(ctx.DbConn)

		lastWordResponse, err := dict.Search(lastWord) // -> optimize (store kana in db on next retrieve)
		if err != nil {
			log.Println(err)
		}
		lastWordKana := lastWordResponse.RelevantKana()

		maybeNextWordResponse, err := dict.Search(maybeNextWord)
		if err != nil {
			log.Println(err)
		}
		if !maybeNextWordResponse.HasEntries() {
			Send(ctx.Bot, "К сожалению, я не знаю такого слова(")
			return
		}

		if maybeNextWordResponse.RelevantSpeechPart() != Noun {
			Send(ctx.Bot, "Слово не является существительным!")
			return
		}
		maybeNextWordKana := maybeNextWordResponse.RelevantKana()

		// Shadow help fix (jisho tries to autocomplete outr words)
		if maybeNextWordResponse.RelevantWord() != maybeNextWord && maybeNextWordKana != maybeNextWord {
			Send(ctx.Bot, "К сожалению, я не знаю такого слова(")
			return
		}

		if IsEnd(maybeNextWordKana) {
			Send(ctx.Bot, "Раунд завершён, введено завершающее слово!")
			ExchangeState("sh_stop") // ??? -> Better state control
			db.ShutDown(ctx.DbConn)
			return
		}

		if db.CheckWordExistence(ctx.DbConn, maybeNextWord) {
			Send(ctx.Bot, "Такое слово уже было")
			return
		}

		if GetLastKana(lastWordKana) == GetFirstKana(maybeNextWordKana) {
			Send(ctx.Bot, fmt.Sprintf("%v, cлово подходит!\n%s「%s」(%s)", ctx.Msg.From.FirstName, maybeNextWordResponse.RelevantWord(), maybeNextWordKana, maybeNextWordResponse.RelevantDefinition()))
			db.AddWord(ctx.DbConn, maybeNextWord, ctx.Msg.From.UserName)
			Send(ctx.Bot, fmt.Sprintf("Следующее слово начинается с:「%c」", GetLastKana(maybeNextWordKana))) // -> what if there is no kanji???, what if we have small kana???
		} else {
			Send(ctx.Bot, "Слово нельзя присоединить(")
			return
		}

	} else {
		Send(ctx.Bot, "Слово не на японском языке!")
		return
	}
}
