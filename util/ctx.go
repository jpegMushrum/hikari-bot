package util

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type MsgContext struct {
	DbConn *gorm.DB
	Bot    *tg.BotAPI
	Msg    *tg.Message
}
