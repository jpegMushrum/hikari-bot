package game

import (
	"log"
	"strings"
)

type GameState uint8

const (
	Init    GameState = 0
	Running GameState = 1
)

var CurrentGameState = Init
var ThreadID int = 0

func Thread() int {
	return ThreadID
}

func SetThreadId(threadId int) {
	ThreadID = threadId
}

func ChangeTo(to GameState) {
	CurrentGameState = to
}

func GetCurrentGameState() GameState {
	return CurrentGameState
}

func IsRunning() bool {
	return GetCurrentGameState() == Running
}

func ExchangeState(command string) (bool, GameState) {
	if atIndex := strings.Index(command, "@"); atIndex != -1 {
		command = command[:atIndex]
	}
	log.Println(command)
	prev := GetCurrentGameState()
	switch command {
	case "/start_game":
		if prev != Init {
			return false, prev
		}
		ChangeTo(Running)
	case "/stop_game":
		if prev == Init {
			return false, prev
		}
		ChangeTo(Init)
	default:
		log.Println("Unexpected game command on state changing!")
	}

	return true, prev
}
