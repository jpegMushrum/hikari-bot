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
	word1 = "名詞「めいし」"
	word2 = "林檎「りんご」"
	word3 = "塩「しお」"
	word4 = "人形「にんぎょう」"
	word5 = "日記「にっき」"
	word6 = "週末「しゅうまつ」"
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
	db.AddPlayer(ctx.DbConn, ctx.Msg.From.UserName)
	Send(ctx.Bot, fmt.Sprintf("%s, добро пожаловать в игру!", ctx.Msg.From.UserName))
}

func PlayerExists(ctx MsgContext) bool {
	return db.CheckPlayerExistence(ctx.DbConn, ctx.Msg.From.UserName)
}

func RunGameCommand(ctx MsgContext) {
	if ok, state := TryChangeState(ctx.Msg.Command()); ok {
		switch state {
		case Init:
			SetChat(ctx.Msg.Chat.ID)
			db.ExecuteScript(ctx.DbConn, db.CreateScript)
			AddPlayer(ctx) // Player who pressed !start
			Send(ctx.Bot, GreetingsString)
			RandomizeStart(ctx)
		case Running:
			Send(ctx.Bot, EndingString)
			// FormAndSendStat(dbConn, bot)
			db.ExecuteScript(ctx.DbConn, db.TruncateScript)
			db.ExecuteScript(ctx.DbConn, db.DeleteScript)
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
		lastWordResponse, err := dict.Search(lastWord) // -> optimize store kanac in db
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

		if GetLastKana(lastWordKana) == GetFirstKana(maybeNextWordKana) {
			Send(ctx.Bot, fmt.Sprintf("%v, cлово подходит!", ctx.Msg.From.UserName))
			db.AddWord(ctx.DbConn, maybeNextWord, ctx.Msg.From.UserName)
			Send(ctx.Bot, fmt.Sprintf("Следующее слово: %s", maybeNextWordKana)) // -> kanji 「kana」
		} else {
			Send(ctx.Bot, "Слово нельзя присоединить(")
			return
		}

	} else {
		Send(ctx.Bot, "Слово не на японском языке!")
		return
	}

	if IsEnd(maybeNextWord) {
		Send(ctx.Bot, "Раунд завершён!")
		TryChangeState("!stop") // ???
		return
	}
}
