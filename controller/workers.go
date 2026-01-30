package controller

import (
	"bakalover/hikari-bot/util"
	"log"
	"sync"
)

type Worker struct {
	ctk     util.ChatThreadKey
	handler Handler
	message chan util.GameContext
	end     chan struct{}
}

func (w *Worker) Run() {
	for {
		select {
		case ctx := <-w.message:
			err := w.handler.Handle(ctx)
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
	mu      sync.Mutex
}

func NewOverseer(handler Handler) *Overseer {
	return &Overseer{
		workers: make(map[util.ChatThreadKey]*Worker),
		handler: handler,
	}
}

func (o *Overseer) GetWorker(ctk util.ChatThreadKey, handler Handler) *Worker {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		return worker
	}

	newWorker := &Worker{
		ctk:     ctk,
		handler: handler,
		message: make(chan util.GameContext, 10),
		end:     make(chan struct{}),
	}

	o.workers[ctk] = newWorker
	go newWorker.Run()

	return newWorker
}

func (o *Overseer) SendMessage(ctx util.GameContext) {
	log.Println("d0")
	worker := o.GetWorker(ctx.CTK, o.handler)

	log.Println("d1")
	worker.message <- ctx
	log.Println("d2")
}

func (o *Overseer) DeleteWorker(ctk util.ChatThreadKey) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if worker, ok := o.workers[ctk]; ok {
		close(worker.end)
		delete(o.workers, ctk)
	}
}
