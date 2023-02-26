/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"net/http"
	"time"

	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/gin-gonic/gin"
)

type GPIOError struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *GPIOError) Error() string {
	if e.Err == "" {
		return "<nil>"
	}
	s := e.Op
	if e.ID != "" {
		s += ":" + e.ID
	}
	s += ": " + e.Err
	return s
}

type GPIO interface {
	ID() string
	Get() (bool, error)
	Configure(config gpio.Config) error
	GetConfig() (gpio.Config, error)
}

type GPIOConfig struct {
	gpio.Config
}

type gpioHandler struct {
	GPIO
	GPIOConfig
}

type GPIOHandler struct {
	io map[string]*gpioHandler
}

func (h *Handler) configGPIO() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(h.GPIO.io) == 0 {
			e := &Error{
				Title:     "Failed to Config GPIO",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		cfg := GPIOConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			e := &Error{
				Title:     "Failed to bind GPIOConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		err := h.GPIO.SetConfig(cfg)
		if err != nil {
			e := &Error{
				Title:     "Failed to SetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		cfg, err = h.GPIO.GetConfig(cfg.ID)
		if err != nil {
			e := &Error{
				Title:     "Failed to GetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}
		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getGPIOS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(h.GPIO.io) == 0 {
			e := &Error{
				Title:     "Failed to GetGPIO",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		gpios, err := h.GPIO.GetConfigAll()
		if len(gpios) == 0 || err != nil {
			notImpl := GPIOError{ID: "", Op: "GetConfigAll", Err: ErrNotImplemented.Error()}
			h.respond(ctx, http.StatusInternalServerError, &notImpl)
			return
		}
		h.respond(ctx, http.StatusOK, gpios)
	}
}

func (g *GPIOHandler) SetConfig(cfg GPIOConfig) error {
	gp, err := g.gpioBy(cfg.ID)
	if err != nil {
		return &GPIOError{ID: cfg.ID, Op: "SetConfig.gpioBy", Err: err.Error()}
	}
	if err := gp.Configure(cfg.Config); err != nil {
		return &GPIOError{ID: cfg.ID, Op: "SetConfig.Configure", Err: err.Error()}
	}
	return nil
}

func (g *GPIOHandler) GetConfigAll() ([]GPIOConfig, error) {
	configs := make([]GPIOConfig, len(g.io))
	pos := 0
	for _, gpi := range g.io {
		cfg, err := gpi.getConfig()
		if err != nil {
			return configs, &GPIOError{ID: cfg.ID, Op: "GetConfigAll.getConfig", Err: err.Error()}
		}
		configs[pos] = cfg
		pos++
	}
	return configs, nil

}
func (g *GPIOHandler) GetConfig(id string) (GPIOConfig, error) {
	gp, err := g.gpioBy(id)
	if err != nil {
		return GPIOConfig{}, &GPIOError{ID: id, Op: "GetConfig.gpioBy", Err: err.Error()}
	}
	return gp.getConfig()
}

func (g *GPIOHandler) gpioBy(id string) (*gpioHandler, error) {
	gp, ok := g.io[id]
	if !ok {
		return nil, ErrNoSuchID
	}
	return gp, nil
}

func (g *gpioHandler) getConfig() (GPIOConfig, error) {
	var err error
	g.GPIOConfig.Config, err = g.GetConfig()
	return g.GPIOConfig, err
}

func (g *GPIOHandler) Open() {
}

func (g *GPIOHandler) Close() {
}
