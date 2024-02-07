package dict

import (
	"testing"
)

const TestFailed = "TEST FAILED"

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
	dict := JishoDict{}
	jr, err := dict.Search(KatakanaOnly)

	if jr.GetKana() != KatakanaOnly || err != nil {
		t.Fatal(TestFailed)
	}

	jr, err = dict.Search(HiraganaOnly)

	if jr.GetKana() != HiraganaOnly || err != nil {
		t.Fatal(TestFailed)
	}
}

func TestBlank(t *testing.T) {
	dict := JishoDict{}
	jr, err := dict.Search(Blank)

	if jr.GetKana() != Blank || err != nil {
		t.Fatal(TestFailed)
	}
}

func TestKanji(t *testing.T) {
	dict := JishoDict{}
	jr, err := dict.Search(Kanji1)

	if jr.GetKana() != KanjiReading1 || err != nil {
		t.Fatal(TestFailed)
	}

	jr, err = dict.Search(Kanji2)

	if jr.GetKana() != KanjiReading2 || err != nil {
		t.Fatal(TestFailed)
	}
}
