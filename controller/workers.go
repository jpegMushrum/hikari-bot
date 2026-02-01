package controller

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
	"fmt"
	"log"
	"sync"

	tele "gopkg.in/telebot.v3"
)

type WorkerContext struct {
	Ctk     util.ChatThreadKey
	TeleCtx tele.Context
	Game    *game.GameState
	Dicts   []dict.Dictionary
	DbConn  *dao.DBConnection
}

type Worker struct {
	ctx     *WorkerContext
	handler Handler
	message chan tele.Context
	end     chan struct{}
}

func (w *Worker) Run() {
	for {
		select {
		case msg := <-w.message:
			w.ctx.TeleCtx = msg

			err := w.handler.Handle(w.ctx)
			if err != nil {
				log.Println("run worker error:\n" + err.Error())
			}
		case <-w.end:
			return
		}
	}
}

type Overseer struct {
	handler Handler
	workers map[util.ChatThreadKey]*Worker
	dicts   []dict.Dictionary
	dsn     string
	mu      sync.Mutex
}

func NewOverseer(handler Handler, dicts []dict.Dictionary, dsn string) *Overseer {
	return &Overseer{
		workers: make(map[util.ChatThreadKey]*Worker),
		handler: handler,
		dicts:   dicts,
		dsn:     dsn,
	}
}

func (o *Overseer) getWorker(ctk util.ChatThreadKey) (*Worker, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		return worker, nil
	}

	dbConn, err := dao.NewConnection(o.dsn)
	if err != nil {
		return nil, fmt.Errorf("get worker controller error: %w", err)
	}

	workerCtx := &WorkerContext{
		Ctk:    ctk,
		Game:   nil,
		Dicts:  o.dicts,
		DbConn: dbConn,
	}

	newWorker := &Worker{
		handler: o.handler,
		ctx:     workerCtx,
		message: make(chan tele.Context, 10),
		end:     make(chan struct{}),
	}

	o.workers[ctk] = newWorker
	go newWorker.Run()

	return newWorker, nil
}

func (o *Overseer) SendMessage(ctx tele.Context) {
	ctk := util.GetCTK(ctx)
	worker, err := o.getWorker(ctk)
	if err != nil {
		log.Println("send message controller error:\n" + err.Error())
		util.Reply(ctx, CallAdmin)
		return
	}

	worker.message <- ctx
}

func (o *Overseer) deleteWorker(ctk util.ChatThreadKey) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		close(worker.end)
		delete(o.workers, ctk)
	}
}
