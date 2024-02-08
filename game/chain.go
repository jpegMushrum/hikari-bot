package game

import (
	"database/sql"

	_ "github.com/lib/pq"
)

const (
	DeadEnd = "ん"
	LongEnd = "ー"
)

func IsNextSuitable(db *sql.DB, word string) bool {

	if !IsJapSuitable(word) {
		return false
	}

	lastSymb := word[len(word)-1]

	if lastSymb == DeadEnd[0] || lastSymb == LongEnd[0] {
		return false
	}

	return true
	// Check small kana on end
	// Get first word from db and check chaining
}
