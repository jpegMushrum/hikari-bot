package controller

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
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
				log.Println(err.Error())
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
	dbConn  *dao.DBConnection
	mu      sync.Mutex
}

func NewOverseer(handler Handler, dicts []dict.Dictionary, dbConn *dao.DBConnection) *Overseer {
	return &Overseer{
		workers: make(map[util.ChatThreadKey]*Worker),
		handler: handler,
		dicts:   dicts,
		dbConn:  dbConn,
	}
}

func (o *Overseer) GetWorker(ctk util.ChatThreadKey, handler Handler) *Worker {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		return worker
	}

	workerCtx := &WorkerContext{
		Ctk:    ctk,
		Game:   nil,
		Dicts:  o.dicts,
		DbConn: o.dbConn,
	}

	newWorker := &Worker{
		handler: handler,
		ctx:     workerCtx,
		message: make(chan tele.Context, 10),
		end:     make(chan struct{}),
	}

	o.workers[ctk] = newWorker
	go newWorker.Run()

	return newWorker
}

func (o *Overseer) SendMessage(ctx tele.Context) {
	ctk := util.GetCTK(ctx)
	worker := o.GetWorker(ctk, o.handler)

	worker.message <- ctx
}

func (o *Overseer) DeleteWorker(ctk util.ChatThreadKey) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		close(worker.end)
		delete(o.workers, ctk)
	}
}
