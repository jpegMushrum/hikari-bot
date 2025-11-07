package util

import (
	"bakalover/hikari-bot/dao"

	tele "gopkg.in/telebot.v3"
)

type GameContext struct {
	TeleCtx tele.Context
	DbConn  *dao.DBConnection
}
