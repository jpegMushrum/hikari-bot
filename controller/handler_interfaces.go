package controller

import (
	"errors"
	"fmt"
)

type Handler interface {
	Handle(*WorkerContext) error
}

type HandlerComposit struct {
	handlers map[string]Handler
}

func NewChain() *HandlerComposit {
	return &HandlerComposit{
		handlers: make(map[string]Handler),
	}
}

func (c *HandlerComposit) AddHandler(trigger string, handler Handler) {
	c.handlers[trigger] = handler
}

func (c *HandlerComposit) Handle(ctx *WorkerContext) error {
	trigger := ctx.TeleCtx.Text()
	handler, ok := c.handlers[trigger]
	if ok {
		return handler.Handle(ctx)
	}

	handler, ok = c.handlers["."]
	if ok {
		return handler.Handle(ctx)
	}

	return errors.New("unknown command handler error: " +
		fmt.Sprintf("%s %v %s", ctx.TeleCtx.Chat().FirstName, ctx.Ctk.ThreadId, ctx.TeleCtx.Sender().FirstName))
}

func NewHandlerComposit() *HandlerComposit {
	return &HandlerComposit{
		handlers: make(map[string]Handler),
	}
}
