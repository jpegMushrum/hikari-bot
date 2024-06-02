package game

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
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

var possibleHiraganaStart = []string{
	"あ", "い", "う", "え", "お",
	"か", "き", "く", "け", "こ",
	"さ", "し", "す", "せ", "そ",
	"た", "ち", "つ", "て", "と",
	"な", "に", "ぬ", "ね", "の",
	"は", "ひ", "ふ", "へ", "ほ",
	"ま", "み", "む", "め", "も",
	"や", "ゆ", "よ",
	"ら", "り", "る", "れ", "ろ",
	"わ",
	"が", "ぎ", "ぐ", "げ", "ご",
	"ざ", "じ", "ず", "ぜ", "ぞ",
	"だ", "ぢ", "づ", "で", "ど",
	"ば", "び", "ぶ", "べ", "ぼ",
	"ぱ", "ぴ", "ぷ", "ぺ", "ぽ",
}

func RandomizeStart(ctx util.GameContext) {
	initKana := possibleHiraganaStart[rand.Intn(len(possibleHiraganaStart))]
	dao.AddWord(ctx.DbConn, initKana, initKana, "DUMMY_USER")
	util.Reply(ctx.TeleCtx, fmt.Sprintf("Первая кана: %s", initKana))
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
	dao.AddWord(ctx.DbConn, word, kana, util.Username(ctx.TeleCtx))
}

func NullifyScore(ctx util.GameContext) {
	dao.SetScore(ctx.DbConn, util.Username(ctx.TeleCtx), 0)
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
	SetThreadId(PoisonedId)
}

func HandleNextWord(ctx util.GameContext, dicts []dict.Dictionary) {
	if !PlayerExists(ctx) {
		AddPlayer(ctx)
	}

	nextWord := ctx.TeleCtx.Text()

	if IsJapSuitable(nextWord) {
		lastWord, lastKana := LastWord(ctx)
		log.Printf("Last Word: %s, %s", lastWord, lastKana)

		nextResponses := make(map[dict.Dictionary]dict.Response)

		// All ops excluding translation performing on first available dict aka Leader Dict
		var leaderDict dict.Dictionary

		isElected := false

		// Different Responses???
		for _, dict := range dicts {
			nextResponse, err := dict.Search(nextWord)
			if err != nil {
				log.Printf("Не удалось найти слово в словаре %v: %v", dict.Repr(), err)
			} else {
				if !isElected {
					leaderDict = dict
				}
				nextResponses[dict] = nextResponse
			}
		}

		if len(nextResponses) == 0 {
			util.Reply(ctx.TeleCtx, "словари недоступны =(")
			return
		}

		nextLeaderResponse := nextResponses[leaderDict]

		nextSpeechParts, err := nextLeaderResponse.RelevantSpeechParts()

		log.Printf("Части речи: %v", nextSpeechParts)

		if err != nil {
			log.Println(err)
			util.Reply(ctx.TeleCtx, err.Error())
			return
		}

		nextKanaSearched, err := nextLeaderResponse.RelevantKana()

		if err != nil {
			log.Println(err)
			util.Reply(ctx.TeleCtx, err.Error())
			return
		}

		nextWordSearched, err := nextLeaderResponse.RelevantWord()

		if err != nil {
			log.Println(err)
			util.Reply(ctx.TeleCtx, err.Error())
			return
		}

		if !HasEntries(nextLeaderResponse) {
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return
		}

		if !ContainsNoun(nextSpeechParts, leaderDict) {
			util.Reply(ctx.TeleCtx, "Слово не является существительным!")
			return
		}

		if IsShadowed(nextWordSearched, nextKanaSearched, nextWord) {
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return
		}

		if IsEnd(nextKanaSearched) {
			util.Reply(ctx.TeleCtx, "Раунд завершён, введено завершающее слово!")
			NullifyScore(ctx)
			ForceStop()
			FormAndSendStats(ctx)
			ClearData(ctx)
			return
		}

		if IsDoubled(ctx, nextWord) {
			util.Reply(ctx.TeleCtx, "Такое слово уже было")
			return
		}

		if GetLastKana(lastKana) == GetFirstKana(nextKanaSearched) {
			wordInfo := fmt.Sprintf("%v, cлово подходит!\n%s「%s」\n-----------------------\n",
				ctx.TeleCtx.Message().Sender.FirstName,
				nextWordSearched,
				nextKanaSearched,
			)

			for dict, nextResponse := range nextResponses {
				nextDefinition, err := nextResponse.RelevantDefinition()
				if err == nil {
					wordInfo += fmt.Sprintf(
						"- %v: %v\n",
						dict.Repr(),
						nextDefinition,
					)
				}
			}

			util.Reply(ctx.TeleCtx, wordInfo)

			AddWord(ctx, nextWordSearched, nextKanaSearched)

			util.Reply(ctx.TeleCtx,
				fmt.Sprintf("Следующее слово начинается с: 「%c」",
					GetLastKana(nextKanaSearched),
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
