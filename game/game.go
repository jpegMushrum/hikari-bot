package game

import (
	"bakalover/hikari-bot/db"
	"bakalover/hikari-bot/dict/jisho"
	"bakalover/hikari-bot/util"
	"fmt"
	"log"
	"math/rand"
	"sort"
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

func RandomizeStart(ctx util.GameContext) {
	words := []string{word1, word2, word3, word4, word5, word6}
	initWord := words[rand.Intn(len(words))]
	db.AddWord(ctx.DbConn, initWord, "DUMMY_USER")
	util.Reply(ctx.TeleCtx, fmt.Sprintf("Первое слово: %s", initWord))
}

func AddPlayer(ctx util.GameContext) {
	db.AddPlayer(ctx.DbConn, util.Username(ctx.TeleCtx), util.FirstName(ctx.TeleCtx))
	util.Reply(ctx.TeleCtx, fmt.Sprintf("%s, добро пожаловать в игру!", util.FirstName(ctx.TeleCtx)))
}

func PlayerExists(ctx util.GameContext) bool {
	return db.CheckPlayerExistence(ctx.DbConn, util.Username(ctx.TeleCtx))
}

func RunGameCommand(ctx util.GameContext) {
	if ok, state := ExchangeState(ctx.TeleCtx.Text()); ok {
		switch state {
		case Init:
			SetThreadId(ctx.TeleCtx.Message().ThreadID)
			db.Init(ctx.DbConn)
			AddPlayer(ctx) // Player who pressed /start_game
			util.Reply(ctx.TeleCtx, GreetingsString)
			RandomizeStart(ctx)
		case Running:
			SetThreadId(-1)
			FormAndSendStats(ctx)
			db.ShutDown(ctx.DbConn)
		}
	} else {
		switch state {
		case Init:
			util.Reply(ctx.TeleCtx, IsNotStartedError)
		case Running:
			util.Reply(ctx.TeleCtx, AlreadyRunningError)
		}
	}
}

func HandleNextWord(ctx util.GameContext, dict *jisho.JishoDict) {
	if !PlayerExists(ctx) {
		AddPlayer(ctx)
	}

	maybeNextWord := ctx.TeleCtx.Text()

	if IsJapSuitable(maybeNextWord) {
		lastWord := db.GetLastWord(ctx.DbConn)

		log.Println("Last Word: ", lastWord)

		lastWordResponse, err := dict.Search(lastWord) // -> optimize (store kana in db on next retrieve)
		if err != nil {
			log.Println(err)
		}
		log.Println("Last word dict response: ", lastWordResponse)
		lastWordKana := lastWordResponse.RelevantKana()

		maybeNextWordResponse, err := dict.Search(maybeNextWord)
		if err != nil {
			log.Println(err)
		}
		if !maybeNextWordResponse.HasEntries() {
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return
		}

		if maybeNextWordResponse.RelevantSpeechPart() != Noun {
			util.Reply(ctx.TeleCtx, "Слово не является существительным!")
			return
		}
		maybeNextWordKana := maybeNextWordResponse.RelevantKana()

		// Shadow help fix (jisho tries to autocomplete outr words)
		if maybeNextWordResponse.RelevantWord() != maybeNextWord && maybeNextWordKana != maybeNextWord {
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return
		}

		if IsEnd(maybeNextWordKana) {
			util.Reply(ctx.TeleCtx, "Раунд завершён, введено завершающее слово!")
			ExchangeState("/stop_game") // ??? -> Better state control
			FormAndSendStats(ctx)
			db.ShutDown(ctx.DbConn)
			return
		}

		if db.CheckWordExistence(ctx.DbConn, maybeNextWord) {
			util.Reply(ctx.TeleCtx, "Такое слово уже было")
			return
		}

		if GetLastKana(lastWordKana) == GetFirstKana(maybeNextWordKana) {
			util.Reply(ctx.TeleCtx, fmt.Sprintf("%v, cлово подходит!\n%s「%s」(%s)", ctx.TeleCtx.Message().Sender.FirstName, maybeNextWordResponse.RelevantWord(), maybeNextWordKana, maybeNextWordResponse.RelevantDefinition()))
			db.AddWord(ctx.DbConn, maybeNextWord, ctx.TeleCtx.Message().Sender.Username)
			util.Reply(ctx.TeleCtx, fmt.Sprintf("Следующее слово начинается с: 「%c」", GetLastKana(maybeNextWordKana)))
		} else {
			util.Reply(ctx.TeleCtx, "Слово нельзя присоединить(")
			return
		}

	} else {
		util.Reply(ctx.TeleCtx, "Слово не на японском языке!")
		return
	}
}

func FormAndSendStats(ctx util.GameContext) {
	players := db.GetAllPlayers(ctx.DbConn)

	// Sort players by score in descending order
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score > players[j].Score
	})

	stats := "Результаты раунда:\n"

	for i, p := range players {
		stats += fmt.Sprintf("%v) %s, Счёт: %v\n", i+1, p.FirstName, p.Score)
	}

	util.Reply(ctx.TeleCtx, stats)
}
