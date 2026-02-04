package util

import (
	"fmt"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func Reply(c tele.Context, what string) {
	c.Bot().Reply(c.Message(), what)
}

func ID(c tele.Context) int64 {
	return c.Message().Sender.ID
}

func Username(c tele.Context) string {
	return c.Message().Sender.Username
}

func FirstName(c tele.Context) string {
	return c.Message().Sender.FirstName
}

func TextWithoutMention(ctx tele.Context) string {
	bot := ctx.Bot()
	text := ctx.Text()

	re := regexp.MustCompile(fmt.Sprintf(`@%s`, regexp.QuoteMeta(bot.Me.Username)))
	return strings.TrimSpace(re.ReplaceAllString(text, ""))
}
