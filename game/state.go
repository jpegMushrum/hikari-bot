package game

import "log"

type GameState uint8

const (
	Init    GameState = 0
	Running GameState = 1
	// Suspend GameState = 2
)

var CurrentGameState = Init

func ChangeTo(to GameState) {
	CurrentGameState = to
}

func GetCurrentGameState() GameState {
	return CurrentGameState
}

// Returns Either (false, PreviousState) or (true, PreviousState)
func TryChangeState(to string) (bool, GameState) {

	prev := GetCurrentGameState()
	switch to {
	case "start":
		if prev != Init {
			return false, prev
		}
		ChangeTo(Running)
	case "stop":
		if prev == Init {
			log.Println("ERR")
			return false, prev
		}
		ChangeTo(Init)
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
