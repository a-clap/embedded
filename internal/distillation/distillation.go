/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
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
	Process        *process.Process
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{
		Engine:  gin.Default(),
		running: atomic.Bool{},
	}

	// Options
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
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
	if h.PTHandler != nil {
		go h.updatePTs()
	}
	err := h.Engine.Run(addr...)
	h.running.Store(false)

	return err
}

func (h *Handler) updatePTs() {
	for h.running.Load() {
		<-time.After(1 * time.Second)
		errs := h.PTHandler.Update()
		if errs != nil {
			log.Debug(errs)
		}
	}
}
