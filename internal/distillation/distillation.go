/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-clap/iot/internal/distillation/process"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	HeatersHandler *HeatersHandler
	DSHandler      *DSHandler
	PTHandler      *PTHandler
	GPIOHandler    *GPIOHandler
	running        atomic.Bool
	runInterval    time.Duration
	finish         chan struct{}
	finished       chan struct{}
	Process        *process.Process
	lastStatus     ProcessStatus
	lastStatusMtx  sync.Mutex
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{
		Engine:        gin.Default(),
		running:       atomic.Bool{},
		finish:        make(chan struct{}),
		finished:      make(chan struct{}),
		runInterval:   1 * time.Second,
		lastStatusMtx: sync.Mutex{},
	}

	// Options
	for _, opt := range opts {
		if err := opt(h); err != nil {
			log.Error(err)
		}
	}
	var err error
	if h.Process, err = process.New(); err != nil {
		panic(err)
	}

	h.routes()

	return h, nil
}

func (h *Handler) Run(addr ...string) error {
	h.running.Store(true)
	go h.updateTemperatures()

	err := h.Engine.Run(addr...)
	h.running.Store(false)
	close(h.finish)
	for range h.finished {
	}

	return err
}

func (h *Handler) updateTemperatures() {
	var wg sync.WaitGroup
	if h.PTHandler != nil {
		wg.Add(1)
		go func() {
			for h.running.Load() {
				select {
				case <-h.finish:
					break
				case <-time.After(h.runInterval):
					errs := h.PTHandler.Update()
					if errs != nil {
						log.Debug(errs)
					}
				}
			}
			wg.Done()
		}()
	}
	if h.DSHandler != nil {
		wg.Add(1)
		go func() {
			for h.running.Load() {
				select {
				case <-h.finish:
					break
				case <-time.After(h.runInterval):
					errs := h.DSHandler.Update()
					if errs != nil {
						log.Debug(errs)
					}
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(h.finish)
}

func (h *Handler) handleProcess() {
	go func() {
		for h.Process.Running() {
			select {
			case <-time.After(h.runInterval):
				s, err := h.Process.Process()
				log.Debug(s)
				if err != nil {
					log.Error(err)
				} else {
					h.updateStatus(s)
				}
			}
		}
	}()

}

func (h *Handler) updateStatus(s process.Status) {
	h.lastStatusMtx.Lock()
	h.lastStatus.Status = s
	h.lastStatusMtx.Unlock()
}
