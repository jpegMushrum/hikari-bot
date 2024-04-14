package util

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func Reply(ctx MsgContext, msg string) {
	reply_msg := tg.NewMessage(ctx.Msg.Chat.ID, msg)
	reply_msg.ReplyToMessageID = ctx.Msg.MessageID
	ctx.Bot.Send(reply_msg)
}
