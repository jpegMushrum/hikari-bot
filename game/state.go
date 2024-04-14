package game

import "log"

type GameState uint8

const (
	Init    GameState = 0
	Running GameState = 1
	// Suspend GameState = 2
)

var CurrentGameState = Init
var GameChatId int64 = 0

func Chat() int64 {
	return int64(GameChatId)
}

func SetChat(chat int64) {
	GameChatId = chat
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

	prev := GetCurrentGameState()
	switch command {
	case "sh_start":
		if prev != Init {
			return false, prev
		}
		ChangeTo(Running)
	case "sh_stop":
		if prev == Init {
			return false, prev
		}
		ChangeTo(Init)
	default:
		log.Fatal("Unexpected game command on state changing!")
		// case "suspend":
		// 	if prev != Running {
		// 		return false, prev
		// 	}
		// 	ChangeTo(Suspend)
		// case "resume":
		// 	if prev != Suspend {
		// 		return false, prev
		// 	}
		// 	ChangeTo(Running)
	}

	return true, prev
}
