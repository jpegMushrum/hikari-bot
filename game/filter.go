package game

import (
	"bakalover/hikari-bot/dict"
	"errors"
	"log"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
	"gopkg.in/telebot.v3"
)

const (
	deadEnd1 = "ん"
	deadEnd2 = "ン"
	longEnd  = "ー"
	noun     = "noun"
)

var smallKana = []rune{
	'ォ', 'ぉ', 'ァ', 'ぁ', 'ゥ', 'ぅ', 'ェ', 'ぇ',
	'ィ', 'ぃ', 'ャ', 'ゃ', 'ョ', 'ょ', 'ュ', 'ゅ',
}

var katakanaToHiragana = map[rune]rune{
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

var smallKanaMappings = map[rune]rune{
	'ォ': 'オ', 'ァ': 'ア', 'ゥ': 'ウ', 'ェ': 'エ', 'ィ': 'イ', 'ャ': 'ヤ', 'ョ': 'ヨ', 'ュ': 'ユ',
	'ぉ': 'お', 'ぁ': 'あ', 'ぅ': 'う', 'ぇ': 'え', 'ぃ': 'い', 'ゃ': 'や', 'ょ': 'よ', 'ゅ': 'ゆ',
}

func toHiragana(kana rune) rune {
	if unicode.In(kana, unicode.Hiragana) {
		return kana
	} else if unicode.In(kana, unicode.Katakana) {
		if converted, ok := katakanaToHiragana[kana]; ok {
			return converted
		}
	}
	log.Println("input is not a hiragana or katakana")
	return 0
}

func isSmall(kana rune) bool {
	for _, char := range smallKana {
		if char == kana {
			return true
		}
	}
	return false
}

func toBigKana(small rune) rune {
	if converted, ok := smallKanaMappings[small]; ok {
		return converted
	} else {
		log.Println("Cannot find small kana to transform")
		return 0
	}
}

func getLastKana(s string) rune {
	var last rune

	for _, r := range s {
		if r == 'ー' {
			last = 'ー'
			continue
		}

		if unicode.In(r, unicode.Hiragana, unicode.Katakana) {
			last = toHiragana(r)
		}
	}

	if isSmall(last) {
		last = toBigKana(last)
	}

	return last
}

func getFirstKana(s string) int32 {
	for _, char := range s {
		if unicode.In(char, unicode.Hiragana, unicode.Katakana) {
			return toHiragana(char)
		}
	}
	return 0
}

func isEnd(word string) bool {
	if strings.HasSuffix(word, deadEnd1) || strings.HasSuffix(word, deadEnd2) || strings.HasSuffix(word, longEnd) {
		return true
	}
	return false
}

func (gs *GameState) isDoubled(word string) bool {
	return gs.dbConn.CheckWordExistence(word)
}

func containsNoun(speechParts []string, dict dict.Dictionary) bool {
	check := false
	for _, s := range speechParts {
		check = check || (s == dict.NounRepr())
	}
	return check
}

func hasEntries(r dict.Response) bool {
	return r.HasEntries()
}

// Shadow help fix (jisho tries to autocomplete our words)
func isShadowed(word1 string, kana1 string, word2 string) bool {
	return word1 != word2 && kana1 != word2
}

func isJapanese(word string) bool {
	for _, char := range word {
		if !unicode.In(char, unicode.Hiragana, unicode.Katakana, unicode.Han) && char != 'ー' {
			return false
		}
	}
	return true
}

func isJapSuitable(word string) bool {
	return len(word) != 0 && isJapanese(word)
}

func (gs *GameState) isTheLastPerson(user *telebot.User) (bool, error) {
	db := gs.dbConn

	lastPlayer := db.LastPlayer()
	if db.Error != nil {
		return false, errors.New("is tha last person error:\n" + db.Error.Error())
	}

	log.Printf("Checking if %v == %v\n", user.ID, lastPlayer)
	return user.ID == int64(lastPlayer), nil
}
