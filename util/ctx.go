package util

import (
	tele "gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

type GameContext struct {
	TeleCtx tele.Context
	DbConn  *gorm.DB
}
