package game

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict/jisho"
	"bakalover/hikari-bot/util"
	"fmt"
	"log"
	"math/rand"
	"sort"
)

const (
	Greetings           = "Раунд начинается!"
	Ending              = "Результаты раунда:"
	IsNotStartedError   = "Игра ещё не началась!"
	AlreadyRunningError = "Игра уже запущена！"
	PoisonedId          = -1
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
	dao.AddWord(ctx.DbConn, initWord, initWord, "DUMMY_USER")
	util.Reply(ctx.TeleCtx, fmt.Sprintf("Первое слово: %s", initWord))
}

func AddPlayer(ctx util.GameContext) {
	dao.AddPlayer(ctx.DbConn, util.Username(ctx.TeleCtx), util.FirstName(ctx.TeleCtx))
	util.Reply(ctx.TeleCtx, fmt.Sprintf("%s, добро пожаловать в игру!", util.FirstName(ctx.TeleCtx)))
}

func PlayerExists(ctx util.GameContext) bool {
	return dao.CheckPlayerExistence(ctx.DbConn, util.Username(ctx.TeleCtx))
}

func InitData(ctx util.GameContext) {
	dao.Init(ctx.DbConn)
}

func ClearData(ctx util.GameContext) {
	dao.ShutDown(ctx.DbConn)
}

func LastWord(ctx util.GameContext) (string, string) {
	return dao.LastWord(ctx.DbConn)
}

// This is bad, really bad
func AllPlayers(ctx util.GameContext) []dao.Player {
	return dao.AllPlayers(ctx.DbConn)
}

func AddWord(ctx util.GameContext, word string, kana string) {
	dao.AddWord(ctx.DbConn, word, kana, ctx.TeleCtx.Message().Sender.Username)
}

func HandleCommand(ctx util.GameContext) {
	state, err := ExchangeState(util.Command(ctx.TeleCtx.Text()))

	if err == nil {
		switch state {
		case Init:
			SetThreadId(ctx.TeleCtx.Message().ThreadID)
			InitData(ctx)
			AddPlayer(ctx) // Player who pressed /start_game
			util.Reply(ctx.TeleCtx, Greetings)
			RandomizeStart(ctx)
		case Running:
			SetThreadId(PoisonedId)
			FormAndSendStats(ctx)
			ClearData(ctx)
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

func ForceStop() {
	ExchangeState(util.StopCommand)
}

func HandleNextWord(ctx util.GameContext, dict *jisho.JishoDict) {
	if !PlayerExists(ctx) {
		AddPlayer(ctx)
	}

	maybeNextWord := ctx.TeleCtx.Text()

	if IsJapSuitable(maybeNextWord) {
		lastWord, lastKana := LastWord(ctx)

		log.Printf("Last Word: %s, %s", lastWord, lastKana)

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

		// Shadow help fix (jisho tries to autocomplete our words)
		if maybeNextWordResponse.RelevantWord() != maybeNextWord && maybeNextWordKana != maybeNextWord {
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return
		}

		if IsEnd(maybeNextWordKana) {
			util.Reply(ctx.TeleCtx, "Раунд завершён, введено завершающее слово!")
			ForceStop()
			FormAndSendStats(ctx)
			ClearData(ctx)
			return
		}

		if IsDoubled(ctx, maybeNextWord) {
			util.Reply(ctx.TeleCtx, "Такое слово уже было")
			return
		}

		if GetLastKana(lastKana) == GetFirstKana(maybeNextWordKana) {
			util.Reply(ctx.TeleCtx,
				fmt.Sprintf("%v, cлово подходит!\n%s「%s」(%s)",
					ctx.TeleCtx.Message().Sender.FirstName,
					maybeNextWordResponse.RelevantWord(),
					maybeNextWordKana,
					maybeNextWordResponse.RelevantDefinition(),
				),
			)

			AddWord(ctx, maybeNextWord, maybeNextWordKana)
			
			util.Reply(ctx.TeleCtx,
				fmt.Sprintf("Следующее слово начинается с: 「%c」",
					GetLastKana(maybeNextWordKana),
				),
			)

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
	players := AllPlayers(ctx)

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
