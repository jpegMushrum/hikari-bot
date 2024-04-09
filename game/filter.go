package game

import (
	"bakalover/hikari-bot/dict"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
)

const (
	DeadEnd = "ん"
	LongEnd = "ー"
	Noun    = "noun"
)

// Broken -> Fix
func GetLastKana(s string) int32 { // -> Check small kana
	for i := len(s) - 1; i >= 0; i-- {
		if unicode.In(rune(s[i]), unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return rune(s[i])
		}
	}
	return 0
}

func GetFirstKana(s string) int32 {
	for _, char := range s {
		if unicode.In(char, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return char
		}
	}
	return 0
}

func IsEnd(word string) bool {
	if strings.HasSuffix(word, DeadEnd) || strings.HasSuffix(word, LongEnd) {
		return true
	}
	return false
}

func IsJapanese(word string) bool {
	for _, char := range word {
		// -> fix "ー" sign
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
