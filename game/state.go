package game

import (
	"bakalover/hikari-bot/util"
	"errors"
	"log"
	"strings"
)

type GameState uint8

const (
	Init    GameState = 0
	Running GameState = 1
)

var CurrentGameState = Init
var ThreadID int = -1

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

func ExchangeState(command util.Command) (GameState, error) {
	if atIndex := strings.Index(string(command), "@"); atIndex != -1 {
		command = command[:atIndex]
	}
	log.Println(command)
	prev := GetCurrentGameState()
	switch command {
	case util.StartCommand:
		if prev != Init {
			return prev, errors.New("")
		}
		ChangeTo(Running)
	case util.StopCommand:
		if prev == Init {
			return prev, errors.New("")
		}
		ChangeTo(Init)
	default:
		log.Println("Unexpected game command on state changing!")
	}

	return prev, nil
}
