package util

type Command string

const (
	StartCommand = Command("/start_game")
	StopCommand  = Command("/stop_game")
	HelpCommand  = Command("/help")
	RulesCommand = Command("/rules")
)
