package game

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/util"
)

type GameState struct {
	ctk           util.ChatThreadKey
	ResultMessage string
	dbConn        *dao.DBConnection
	dicts         []dict.Dictionary
}

func NewGame(ctk util.ChatThreadKey, dbConn *dao.DBConnection, dicts []dict.Dictionary) *GameState {
	return &GameState{
		ctk:    ctk,
		dbConn: dbConn,
		dicts:  dicts,
	}
}

func (gs *GameState) Thread() util.ChatThreadKey {
	return gs.ctk
}

type WordHandleResult int

const (
	Success           WordHandleResult = 0
	GotError          WordHandleResult = 1
	FoundLastPerson   WordHandleResult = 2
	WordNotJapanese   WordHandleResult = 3
	DictsNotAnswering WordHandleResult = 4
	NoSpeachPart      WordHandleResult = 5
	NoSuchWord        WordHandleResult = 6
	GotEndWord        WordHandleResult = 7
	GotDoubledWord    WordHandleResult = 8
	CantJoinWords     WordHandleResult = 9
)
