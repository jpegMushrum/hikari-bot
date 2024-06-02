package jisho

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	KatakanaOnly  = "スキー"
	HiraganaOnly  = "はらきり"
	Blank         = ""
	Kanji1        = "太陽"
	KanjiReading1 = "たいよう"
	Kanji2        = "気候変動枠組条約締約国会議"
	KanjiReading2 = "きこうへんどうわくぐみじょうやくていやくこくかいぎ"
)

func TestKana(t *testing.T) {
	dict := Jisho{}
	jr, _ := dict.Search(KatakanaOnly)
	assert.True(t, jr.HasEntries())
	kana, _ := jr.RelevantKana()
	assert.True(t, kana == KatakanaOnly)

	jr, _ = dict.Search(HiraganaOnly)
	assert.True(t, jr.HasEntries())
	kana, _ = jr.RelevantKana()
	assert.True(t, kana == HiraganaOnly)
}

func TestBlank(t *testing.T) {
	dict := Jisho{}

	jr, _ := dict.Search(Blank)
	assert.False(t, jr.HasEntries())
}

func TestKanji(t *testing.T) {
	dict := Jisho{}

	jr, _ := dict.Search(Kanji1)
	assert.True(t, jr.HasEntries())
	kana, _ := jr.RelevantKana()
	assert.True(t, kana == KanjiReading1)

	jr, _ = dict.Search(Kanji2)
	assert.True(t, jr.HasEntries())
	kana, _ = jr.RelevantKana()
	assert.True(t, kana == KanjiReading2)
}
