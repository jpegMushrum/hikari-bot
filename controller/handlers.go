package controller

import (
	"bakalover/hikari-bot/util"
	"errors"
	"log"
)

type Handler interface {
	Handle(util.GameContext) error
}

type SimpleHandler struct {
	Inner func(util.GameContext) error
}

func (h *SimpleHandler) Handle(ctx util.GameContext) error {
	return h.Inner(ctx)
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

func (c *HandlerComposit) Handle(ctx util.GameContext) error {
	trigger := ctx.TeleCtx.Text()
	log.Println("d|")

	handler, ok := c.handlers[trigger]
	if ok {
		return handler.Handle(ctx)
	}

	handler, ok = c.handlers["."]
	if !ok {
		return errors.New("Don't have handler for this command")
	}

	return handler.Handle(ctx)
}

func NewHandlerComposit() *HandlerComposit {
	return &HandlerComposit{
		handlers: make(map[string]Handler),
	}
}
