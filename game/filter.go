package game

import (
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/util"
	"log"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
	"gopkg.in/telebot.v3"
)

const (
	DeadEnd1 = "ん"
	DeadEnd2 = "ン"
	LongEnd  = "ー"
	Noun     = "noun"
)

var SmallKana = []rune{
	'ォ', 'ぉ', 'ァ', 'ぁ', 'ゥ', 'ぅ', 'ェ', 'ぇ',
	'ィ', 'ぃ', 'ャ', 'ゃ', 'ョ', 'ょ', 'ュ', 'ゅ',
}

var KatakanaToHiragana = map[rune]rune{
	// Vowels
	'ア': 'あ', 'イ': 'い', 'ウ': 'う', 'エ': 'え', 'オ': 'お',
	// Vowels with diacritics
	'ァ': 'ぁ', 'ィ': 'ぃ', 'ゥ': 'ぅ', 'ェ': 'ぇ', 'ォ': 'ぉ',
	'ヴ': 'ゔ',
	// Consonants
	'カ': 'か', 'キ': 'き', 'ク': 'く', 'ケ': 'け', 'コ': 'こ',
	'サ': 'さ', 'シ': 'し', 'ス': 'す', 'セ': 'せ', 'ソ': 'そ',
	'タ': 'た', 'チ': 'ち', 'ツ': 'つ', 'テ': 'て', 'ト': 'と',
	'ナ': 'な', 'ニ': 'に', 'ヌ': 'ぬ', 'ネ': 'ね', 'ノ': 'の',
	'ハ': 'は', 'ヒ': 'ひ', 'フ': 'ふ', 'ヘ': 'へ', 'ホ': 'ほ',
	'マ': 'ま', 'ミ': 'み', 'ム': 'む', 'メ': 'め', 'モ': 'も',
	'ヤ': 'や', 'ユ': 'ゆ', 'ヨ': 'よ',
	'ラ': 'ら', 'リ': 'り', 'ル': 'る', 'レ': 'れ', 'ロ': 'ろ',
	'ワ': 'わ', 'ヲ': 'を', 'ン': 'ん',
	'ガ': 'が', 'ギ': 'ぎ', 'グ': 'ぐ', 'ゲ': 'げ', 'ゴ': 'ご',
	'ザ': 'ざ', 'ジ': 'じ', 'ズ': 'ず', 'ゼ': 'ぜ', 'ゾ': 'ぞ',
	'ダ': 'だ', 'ヂ': 'ぢ', 'ヅ': 'づ', 'デ': 'で', 'ド': 'ど',
	'バ': 'ば', 'ビ': 'び', 'ブ': 'ぶ', 'ベ': 'べ', 'ボ': 'ぼ',
	'パ': 'ぱ', 'ピ': 'ぴ', 'プ': 'ぷ', 'ペ': 'ぺ', 'ポ': 'ぽ',
}

var SmallKanaMappings = map[rune]rune{
	'ォ': 'オ', 'ァ': 'ア', 'ゥ': 'ウ', 'ェ': 'エ', 'ィ': 'イ', 'ャ': 'ヤ', 'ョ': 'ヨ', 'ュ': 'ユ',
	'ぉ': 'お', 'ぁ': 'あ', 'ぅ': 'う', 'ぇ': 'え', 'ぃ': 'い', 'ゃ': 'や', 'ょ': 'よ', 'ゅ': 'ゆ',
}

func ToHiragana(kana rune) rune {
	if unicode.In(kana, unicode.Hiragana) {
		return kana
	} else if unicode.In(kana, unicode.Katakana) {
		if converted, ok := KatakanaToHiragana[kana]; ok {
			return converted
		}
	}
	log.Println("input is not a hiragana or katakana")
	return 0
}

func IsSmall(kana rune) bool {
	for _, char := range SmallKana {
		if char == kana {
			return true
		}
	}
	return false
}

func ToBigKana(small rune) rune {
	if converted, ok := SmallKanaMappings[small]; ok {
		return converted
	} else {
		log.Println("Cannot find small kana to transform")
		return 0
	}
}

func GetLastKana(s string) rune {
	var last rune

	for _, r := range s {
		if r == 'ー' {
			last = 'ー'
			continue
		}

		if unicode.In(r, unicode.Hiragana, unicode.Katakana) {
			last = ToHiragana(r)
		}
	}

	if IsSmall(last) {
		last = ToBigKana(last)
	}

	return last
}

func GetFirstKana(s string) int32 {
	for _, char := range s {
		if unicode.In(char, unicode.Hiragana, unicode.Katakana) {
			return ToHiragana(char)
		}
	}
	return 0
}

func IsEnd(word string) bool {
	if strings.HasSuffix(word, DeadEnd1) || strings.HasSuffix(word, DeadEnd2) || strings.HasSuffix(word, LongEnd) {
		return true
	}
	return false
}

func IsDoubled(ctx util.GameContext, word string) bool {
	return ctx.DbConn.CheckWordExistence(word)
}

func ContainsNoun(speechParts []string, dict dict.Dictionary) bool {
	check := false
	for _, s := range speechParts {
		check = check || (s == dict.NounRepr())
	}
	return check
}

func HasEntries(r dict.Response) bool {
	return r.HasEntries()
}

// Shadow help fix (jisho tries to autocomplete our words)
func IsShadowed(word1 string, kana1 string, word2 string) bool {
	return word1 != word2 && kana1 != word2
}

func IsJapanese(word string) bool {
	for _, char := range word {
		if !unicode.In(char, unicode.Hiragana, unicode.Katakana, unicode.Han) && char != 'ー' {
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

func IsTheLastPerson(user *telebot.User, ctx util.GameContext) bool {
	lastPlayer := ctx.DbConn.LastPlayer()
	log.Printf("Checking if %v == %v\n", user.ID, lastPlayer)
	return user.ID == int64(lastPlayer)
}
