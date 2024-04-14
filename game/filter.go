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

var SmallKana = []rune{
	'ォ', 'ぉ', 'ァ', 'ぁ', 'ゥ', 'ぅ', 'ェ',
	'ぇ', 'ィ', 'ぃ', 'ャ', 'ゃ', 'ョ', 'ょ',
}

func IsSmall(kana rune) bool {
	for _, char := range SmallKana {
		if char == kana {
			return true
		}
	}
	return false
}

func MapSmallToBig(kana rune) rune {
	switch kana {
	case 'ォ':
		return 'オ'
	case 'ぉ':
		return 'お'

	case 'ァ':
		return 'ア'

	case 'ぁ':
		return 'あ'

	case 'ゥ':
		return 'ウ'

	case 'ぅ':
		return 'う'

	case 'ェ':
		return 'エ'

	case 'ぇ':
		return 'え'

	case 'ィ':
		return 'イ'

	case 'ぃ':
		return 'い'

	case 'ャ':
		return 'ヤ'

	case 'ゃ':
		return 'や'

	case 'ョ':
		return 'ヨ'

	case 'ょ':
		return 'よ'

	}
	return 0
}

func GetLastKana(s string) int32 {
	var ans rune = 0

outter_loop:
	for i := len(s) - 1; i >= 0; i-- {
		for _, char := range s[i:] { // Oh shit
			if char == 'ー' { // Doesn't work with Range tables
				return 'ー'
			}
			if unicode.In(char, unicode.Hiragana, unicode.Katakana, unicode.Han) {
				ans = char
				break outter_loop
			}
		}
	}

	if IsSmall(ans) {
		ans = MapSmallToBig(ans)
	}
	return ans
}

func GetFirstKana(s string) int32 {
	for _, char := range s {
		if char == 'ー' { // Doesn't work with Range tables
			return 'ー'
		}
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
