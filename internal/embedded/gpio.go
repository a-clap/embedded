/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"net/http"

	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/gin-gonic/gin"
)

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
	gpios map[string]*gpioHandler
}

var (
	ErrNoSuchGPIO = errors.New("specified input doesnt' exist")
)

func (h *Handler) configGPIO() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := GPIOConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		err := h.GPIO.SetConfig(cfg)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		cfg, err = h.GPIO.GetConfig(cfg.ID)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}
		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getGPIOS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gpios, err := h.GPIO.GetConfigAll()
		if len(gpios) == 0 || err != nil {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, gpios)
	}
}

func (g *GPIOHandler) SetConfig(cfg GPIOConfig) error {
	gp, err := g.gpioBy(cfg.ID)
	if err != nil {
		return err
	}

	return gp.Configure(cfg.Config)
}

func (g *GPIOHandler) GetConfigAll() ([]GPIOConfig, error) {
	configs := make([]GPIOConfig, len(g.gpios))
	pos := 0
	for _, gpi := range g.gpios {
		cfg, err := gpi.getConfig()
		if err != nil {
			return configs, err
		}
		configs[pos] = cfg
		pos++
	}
	return configs, nil

}
func (g *GPIOHandler) GetConfig(hwid string) (GPIOConfig, error) {
	gp, err := g.gpioBy(hwid)
	if err != nil {
		return GPIOConfig{}, err
	}
	return gp.getConfig()
}

func (g *GPIOHandler) gpioBy(hwid string) (*gpioHandler, error) {
	gp, ok := g.gpios[hwid]
	if !ok {
		return nil, ErrNoSuchGPIO
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
