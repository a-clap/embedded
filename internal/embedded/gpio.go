/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GPIOHandler struct {
	gpios map[string]models.GPIO
}

var (
	ErrNoSuchGPIO = errors.New("specified input doesnt' exist")
)

func (h *Handler) configGPIO() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := models.GPIOConfig{}
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

func (g *GPIOHandler) SetConfig(cfg models.GPIOConfig) error {
	gpio, err := g.gpioBy(cfg.ID)
	if err != nil {
		return err
	}
	return gpio.SetConfig(cfg)
}

func (g *GPIOHandler) GetConfigAll() ([]models.GPIOConfig, error) {
	configs := make([]models.GPIOConfig, len(g.gpios))
	pos := 0
	for _, gpio := range g.gpios {
		cfg, err := gpio.Config()
		if err != nil {
			return configs, err
		}
		configs[pos] = cfg
		pos++
	}
	return configs, nil

}
func (g *GPIOHandler) GetConfig(hwid string) (models.GPIOConfig, error) {
	gpio, err := g.gpioBy(hwid)
	if err != nil {
		return models.GPIOConfig{}, err
	}
	return gpio.Config()
}

func (g *GPIOHandler) gpioBy(hwid string) (models.GPIO, error) {
	gpio, ok := g.gpios[hwid]
	if !ok {
		return nil, ErrNoSuchGPIO
	}
	return gpio, nil
}

func (g *GPIOHandler) Open() {
}

func (g *GPIOHandler) Close() {
}
