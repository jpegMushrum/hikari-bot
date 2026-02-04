package controller

import (
	"bakalover/hikari-bot/dao"
	"bakalover/hikari-bot/dict"
	"bakalover/hikari-bot/game"
	"bakalover/hikari-bot/util"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
)

const Timeout = 5 * time.Minute

type WorkerContext struct {
	CTK     util.ChatThreadKey
	TeleCtx tele.Context
	Game    *game.GameState
	Dicts   []dict.Dictionary
	DbConn  *dao.DBConnection

	backCtx   context.Context
	cancel    context.CancelFunc
	keepAlive chan struct{}
	stopTimer chan struct{}
}

type Worker struct {
	ctx     *WorkerContext
	handler Handler
	message chan tele.Context
	end     chan struct{}
}

func (w *WorkerContext) StartTTLWatcher() {
	w.keepAlive = make(chan struct{}, 1)
	w.stopTimer = make(chan struct{}, 1)

	w.backCtx, w.cancel = context.WithCancel(context.Background())

	go func() {
		timer := time.NewTimer(Timeout)
		defer timer.Stop()

		for {
			select {
			case <-w.backCtx.Done():
				return
			case <-w.keepAlive:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(Timeout)

			case <-w.stopTimer:
				if !timer.Stop() {
					<-timer.C
				}

			case <-timer.C:
				w.cancel()
				return
			}
		}
	}()
}

func (w *WorkerContext) RefreshTTL() {
	select {
	case w.keepAlive <- struct{}{}:
	default:
	}
}

func (w *WorkerContext) StopTTL() {
	select {
	case w.stopTimer <- struct{}{}:
	default:
	}
}

func (w *Worker) Run(die func(util.ChatThreadKey)) {
	for {
		select {
		case msg := <-w.message:
			w.ctx.TeleCtx = msg

			err := w.handler.Handle(w.ctx)
			if err != nil {
				util.Reply(w.ctx.TeleCtx, CallAdmin)
				log.Println("run worker error:\n" + err.Error())
			}
		case <-w.ctx.backCtx.Done():
			die(w.ctx.CTK)
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
		CTK:    ctk,
		Game:   nil,
		Dicts:  o.dicts,
		DbConn: dbConn,
	}

	workerCtx.StartTTLWatcher()

	newWorker := &Worker{
		handler: o.handler,
		ctx:     workerCtx,
		message: make(chan tele.Context, 10),
		end:     make(chan struct{}),
	}

	o.workers[ctk] = newWorker
	go newWorker.Run(o.deleteWorker)

	return newWorker, nil
}

func (o *Overseer) SendMessage(ctx tele.Context) {
	ctx.Message().Text = util.TextWithoutMention(ctx)
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
		worker.ctx.cancel()
		delete(o.workers, ctk)
	}
}
