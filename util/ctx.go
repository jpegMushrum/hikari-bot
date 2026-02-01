package util

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"

	tele "gopkg.in/telebot.v3"
)

type ChatThreadKey struct {
	ChatId   int64
	ThreadId int
}

func GetCTK(ctx tele.Context) ChatThreadKey {
	return ChatThreadKey{
		ChatId:   ctx.Chat().ID,
		ThreadId: ctx.Message().ThreadID,
	}
}

type WorkerContext struct {
	CTK     ChatThreadKey
	Dicts   []dict.Dictionary
	TeleCtx tele.Context
	DbConn  *dao.DBConnection
}

func GetGameKey(ctk ChatThreadKey) dao.GameKey {
	return dao.GameKey{
		ChatID:   ctk.ChatId,
		ThreadID: ctk.ThreadId,
	}
}
