package game

import (
	"bakalover/hikari-bot/dict"
	"database/sql"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
)

const (
	DeadEnd = "ん"
	LongEnd = "ー"
	Noun    = "noun"
)

func GetLastKana(s string) int32 {
	for i := len(s) - 1; i >= 0; i-- {
		if unicode.In(rune(s[i]), unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return rune(s[i])
		}
	}
	return 0
}

// Checks if word is bassically in Japanese and have suitable ending
func IsNextSuitable(db *sql.DB, word string) bool {

	if !IsJapSuitable(word) {
		return false
	}

	if strings.HasSuffix(word, DeadEnd) || strings.HasSuffix(word, LongEnd) {
		return false
	}

	return true
	// Check small kana on end
	// Get first word from db and check chaining
}

func IsJapanese(word string) bool {
	for _, char := range word {
		if !unicode.In(char, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return false
		}
	}
	return true
}

func IsNotBlank(word string) bool {
	return len(word) != 0
}

func IsJapSuitable(word string) bool {
	return IsNotBlank(word) && IsJapanese(word)
}

func IsNoun(jsr dict.Response) bool {
	return jsr.RelevantSpeechPart() == Noun
}
