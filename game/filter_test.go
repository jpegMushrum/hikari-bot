package game

import "testing"

func TestKana(t *testing.T) {
	if GetFirstKana("へんたい") != 'へ' {
		t.Fatal()
	}
	if GetFirstKana("キス") != 'キ' {
		t.Fatal()
	}
}
