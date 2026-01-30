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
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	initKana := possibleHiraganaStart[rand.Intn(len(possibleHiraganaStart))]
	db.AddWord(initKana, initKana, "DUMMY_USER", 0)
	util.Reply(ctx.TeleCtx, fmt.Sprintf("Первая кана: %s", initKana))
}

func AddPlayer(ctx util.GameContext) {
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	db.AddPlayer(util.ID(ctx.TeleCtx), util.Username(ctx.TeleCtx), util.FirstName(ctx.TeleCtx))
	util.Reply(ctx.TeleCtx, fmt.Sprintf("%s, добро пожаловать в игру!", util.FirstName(ctx.TeleCtx)))
}

func PlayerExists(ctx util.GameContext) bool {
	db := ctx.DbConn
	if db.Error != nil {
		return false
	}

	return db.CheckPlayerExistence(util.Username(ctx.TeleCtx))
}

func InitData(ctx util.GameContext) {
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	db.Init()
}

func ClearData(ctx util.GameContext) {
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	db.ClearTables()
}

func LastWord(ctx util.GameContext) (string, string) {
	db := ctx.DbConn
	if db.Error != nil {
		return "", ""
	}

	return db.LastWord()
}

func AllPlayers(ctx util.GameContext) []dao.Player {
	db := ctx.DbConn
	if db.Error != nil {
		return nil
	}

	return db.AllPlayers()
}

func AddWord(ctx util.GameContext, word string, kana string) {
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	db.AddWord(word, kana, util.Username(ctx.TeleCtx), util.ID(ctx.TeleCtx))
}

func NullifyScore(ctx util.GameContext) {
	db := ctx.DbConn
	if db.Error != nil {
		return
	}

	db.SetScore(util.Username(ctx.TeleCtx), 0)
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

func HandleNextWord(ctx util.GameContext) {
	dicts := ctx.Dicts
	ctx.DbConn.Reset()

	if !PlayerExists(ctx) {
		AddPlayer(ctx)
	}

	if ctx.DbConn.Error != nil {
		util.Reply(ctx.TeleCtx, "Бот упал, обратитесь к админу!")
		return
	}

	nextWord := ctx.TeleCtx.Text()
	nextPerson := ctx.TeleCtx.Sender()

	if IsTheLastPerson(nextPerson, ctx) {
		util.Reply(ctx.TeleCtx, fmt.Sprintf("Неправильная очередь, %s добавил прошлое слово!", nextPerson.FirstName))
		return
	}

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
					isElected = true
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

		switch {
		case ctx.DbConn.Error != nil:
			util.Reply(ctx.TeleCtx, "Бот упал, обратитесь к админу!")
			return

		case !HasEntries(nextLeaderResponse):
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return

		case !IsJapanese(nextWordSearched):
			util.Reply(ctx.TeleCtx, "Слово не на японском языке!")
			return

		case !ContainsNoun(nextSpeechParts, leaderDict):
			util.Reply(ctx.TeleCtx, "Слово не является существительным!")
			return

		case IsShadowed(nextWordSearched, nextKanaSearched, nextWord):
			util.Reply(ctx.TeleCtx, "К сожалению, я не знаю такого слова(")
			return

		case IsEnd(nextKanaSearched):
			wordInfo := WordInfo(ctx, "Слово не подходит", nextWordSearched, nextKanaSearched, nextResponses)
			util.Reply(ctx.TeleCtx, "Раунд завершён, введено завершающее слово!\n"+wordInfo)
			NullifyScore(ctx)
			ForceStop()
			FormAndSendStats(ctx)
			ClearData(ctx)
			return

		case IsDoubled(ctx, nextWordSearched):
			util.Reply(ctx.TeleCtx, "Такое слово уже было")
			return

		case GetLastKana(lastKana) != GetFirstKana(nextKanaSearched):
			util.Reply(ctx.TeleCtx, "Слово нельзя присоединить(")
			return

		case GetLastKana(lastKana) == GetFirstKana(nextKanaSearched):
			wordInfo := WordInfo(ctx, "Слово подходит", nextWordSearched, nextKanaSearched, nextResponses)

			util.Reply(ctx.TeleCtx, wordInfo)
			AddWord(ctx, nextWordSearched, nextKanaSearched)
			util.Reply(ctx.TeleCtx,
				fmt.Sprintf("Следующее слово начинается с: 「%c」",
					GetLastKana(nextKanaSearched),
				),
			)

		default:
			util.Reply(ctx.TeleCtx, "Неизвестная ошибка, обратитесь к админу!")
		}
	} else {
		util.Reply(ctx.TeleCtx, "Слово не на японском языке!")
	}

	if ctx.DbConn.Error != nil {
		log.Println(ctx.DbConn.Error)
	}
}

func WordInfo(ctx util.GameContext, msg, word, kana string, responses map[dict.Dictionary]dict.Response) string {
	wordInfo := fmt.Sprintf("%v, %v!\n%s「%s」\n-----------------------\n",
		msg,
		ctx.TeleCtx.Message().Sender.FirstName,
		word,
		kana,
	)

	for dict, nextResponse := range responses {
		nextDefinition, err := nextResponse.RelevantDefinition()
		if err == nil {
			wordInfo += fmt.Sprintf("- %v: %v\n", dict.Repr(), nextDefinition)
		}
	}

	return wordInfo
}

func FormAndSendStats(ctx util.GameContext) {
	db := ctx.DbConn
	players := db.AllPlayers()

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
