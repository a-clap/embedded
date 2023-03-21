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
	updateInterval time.Duration
	finish         chan struct{}
	finished       chan struct{}
	Process        *process.Process
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{
		Engine:         gin.Default(),
		running:        atomic.Bool{},
		finish:         make(chan struct{}),
		finished:       make(chan struct{}),
		updateInterval: 1 * time.Second,
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
				case <-time.After(h.updateInterval):
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
				case <-time.After(h.updateInterval):
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
