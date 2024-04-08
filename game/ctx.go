package game

import (
	"database/sql"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MsgContext struct {
	DbConn *sql.DB
	Bot    *tg.BotAPI
	Msg    *tg.Message
}
