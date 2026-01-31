package game

import (
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/util"
	"fmt"
	"log"
	"math/rand"
	"sort"

	tele "gopkg.in/telebot.v3"
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

func (gs *GameState) randomizeStart() (string, error) {
	db := gs.dbConn

	initKana := possibleHiraganaStart[rand.Intn(len(possibleHiraganaStart))]

	db.AddWord(initKana, initKana, "DUMMY_USER", 0)
	if db.Error != nil {
		return "", fmt.Errorf("randomize start error: %w", db.Error)
	}

	return initKana, nil
}

func (gs *GameState) addPlayer(ctx tele.Context) error {
	db := gs.dbConn

	db.AddPlayer(util.ID(ctx), util.Username(ctx), util.FirstName(ctx))
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func (gs *GameState) playerExists(ctx tele.Context) (bool, error) {
	db := gs.dbConn

	result := db.CheckPlayerExistence(util.Username(ctx))
	if db.Error != nil {
		return false, db.Error
	}

	return result, nil
}

func (gs *GameState) lastWord() (string, string, error) {
	db := gs.dbConn

	word, read := db.LastWord()
	if db.Error != nil {
		return "", "", fmt.Errorf("last word game error: %w", db.Error)
	}

	return word, read, nil
}

func (gs *GameState) addWord(ctx tele.Context, word string, kana string) error {
	db := gs.dbConn

	db.AddWord(word, kana, util.Username(ctx), util.ID(ctx))
	if db.Error != nil {
		return fmt.Errorf("add word game error: %w", db.Error)
	}

	return nil
}

func (gs *GameState) StartGame() (string, error) {
	db := gs.dbConn
	db.Reset()

	db.Init()
	if db.Error != nil {
		return "", fmt.Errorf("start game error: %w", db.Error)
	}

	return gs.randomizeStart()
}

func (gs *GameState) StopGame() error {
	db := gs.dbConn
	db.Reset()

	db.ClearTables()
	if db.Error != nil {
		return fmt.Errorf("stop game error: %w", db.Error)
	}

	return nil
}

func (gs *GameState) HandleNextWord(ctx tele.Context) (WordHandleResult, error) {
	dicts := gs.dicts
	db := gs.dbConn
	db.Reset()

	if !isJapSuitable(ctx.Text()) {
		return WordNotJapanese, nil
	}

	ok, err := gs.playerExists(ctx)
	if err != nil {
		return GotError, fmt.Errorf("handle next word game error: %w", err)
	}
	if !ok {
		err := gs.addPlayer(ctx)
		if err != nil {
			return GotError, fmt.Errorf("handle next word game error: %w", err)
		}
	}

	nextPerson := ctx.Sender()

	ok, err = gs.isTheLastPerson(nextPerson)
	if err != nil {
		return GotError, fmt.Errorf("handle next word game error: %w", err)
	}
	if ok {
		return FoundLastPerson, nil
	}

	nextWord := ctx.Text()

	if !isJapSuitable(nextWord) {
		return WordNotJapanese, nil
	}

	lastWord, lastKana, err := gs.lastWord()
	if err != nil {
		return GotError, fmt.Errorf("handle next word game error: %w", err)
	}

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
		return DictsNotAnswering, nil
	}

	nextLeaderResponse := nextResponses[leaderDict]
	nextSpeechParts, err := nextLeaderResponse.RelevantSpeechParts()

	if err != nil {
		return NoSpeachPart, nil
	}

	log.Printf("Части речи: %v", nextSpeechParts)

	nextKanaSearched, err := nextLeaderResponse.RelevantKana()
	if err != nil {
		return NoSuchWord, nil
	}

	nextWordSearched, err := nextLeaderResponse.RelevantWord()
	if err != nil {
		return NoSuchWord, nil
	}

	ok = gs.isDoubled(nextWordSearched)
	if db.Error != nil {
		return GotError, fmt.Errorf("handle next word game error: %w", db.Error)
	}
	if ok {
		return GotDoubledWord, nil
	}

	switch {
	case db.Error != nil:
		return GotError, fmt.Errorf("handle next word game error: %w", db.Error)

	case isShadowed(nextWordSearched, nextKanaSearched, nextWord):
		return NoSuchWord, nil

	case !hasEntries(nextLeaderResponse):
		return NoSuchWord, nil

	case !isJapanese(nextWordSearched):
		return WordNotJapanese, nil

	case !containsNoun(nextSpeechParts, leaderDict):
		return NoSpeachPart, nil

	case getLastKana(lastKana) != getFirstKana(nextKanaSearched):
		return CantJoinWords, nil

	case isEnd(nextKanaSearched):
		wordInfo := wordInfo(ctx, "Слово не подходит", nextWordSearched, nextKanaSearched, nextResponses)
		gs.ResultMessage = "Раунд завершён, введено завершающее слово!\n\n" + wordInfo

		return GotEndWord, nil
	default:
	}

	wordInfo := wordInfo(ctx, "Слово подходит", nextWordSearched, nextKanaSearched, nextResponses)
	wordInfo += fmt.Sprintf("\nСледующее слово начинается с: 「%c」", getLastKana(nextKanaSearched))
	gs.ResultMessage = wordInfo

	err = gs.addWord(ctx, nextWordSearched, nextKanaSearched)
	if err != nil {
		return GotError, fmt.Errorf("handle next word game error: %w", err)
	}

	return Success, nil
}

func wordInfo(ctx tele.Context, msg, word, kana string, responses map[dict.Dictionary]dict.Response) string {
	wordInfo := fmt.Sprintf("%v, %v!\n%s「%s」\n-----------------------\n",
		msg,
		ctx.Message().Sender.FirstName,
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

func (gs *GameState) FormStats() (string, error) {
	db := gs.dbConn
	players := db.AllPlayers()

	if db.Error != nil {
		return "", fmt.Errorf("form stats game error: %w", db.Error)
	}

	// Sort players by score in descending order
	sort.Slice(players, func(i, j int) bool {
		return players[i].Score > players[j].Score
	})

	stats := "Результаты раунда:\n"
	for i, p := range players {
		stats += fmt.Sprintf("%v) %s, Счёт: %v\n", i+1, p.FirstName, p.Score)
	}

	return stats, nil
}
