/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrHeaterDoesntExist = errors.New("heater doesn't exist")
)

type HeaterError struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *HeaterError) Error() string {
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

type Heater interface {
	Enable(ena bool)
	SetPower(pwr uint) error
	Enabled() bool
	Power() uint
}

type HeaterConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Power   uint   `json:"power"`
}

type HeaterHandler struct {
	heaters map[string]Heater
}

func (h *Handler) configHeater() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		if err := h.Heaters.Config(cfg); err != nil {
			if heaterError, ok := err.(*HeaterError); ok {
				h.respond(ctx, http.StatusInternalServerError, heaterError)
			} else {
				h.respond(ctx, http.StatusInternalServerError, toError(err))
			}
			return
		}

		s, _ := h.Heaters.StatusBy(cfg.ID)
		h.respond(ctx, http.StatusOK, s)
	}
}

func (h *Handler) getHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		heaters := h.Heaters.Status()
		if len(heaters) == 0 {
			notImpl := HeaterError{ID: "", Op: "GetTemperatures", Err: ErrNotImplemented.Error()}
			h.respond(ctx, http.StatusInternalServerError, &notImpl)
			return
		}
		h.respond(ctx, http.StatusOK, heaters)
	}
}

func (h *HeaterHandler) Config(cfg HeaterConfig) error {
	heater, err := h.by(cfg.ID)
	if err != nil {
		return &HeaterError{ID: cfg.ID, Op: "Config", Err: err.Error()}
	}

	if err := heater.SetPower(cfg.Power); err != nil {
		return err
	}
	heater.Enable(cfg.Enabled)
	return nil
}

func (h *HeaterHandler) Enable(id string, ena bool) error {
	heat, err := h.by(id)
	if err != nil {
		return &HeaterError{ID: id, Op: "Enable", Err: err.Error()}
	}

	heat.Enable(ena)
	return nil
}

func (h *HeaterHandler) Power(id string, pwr uint) error {
	heat, err := h.by(id)
	if err != nil {
		return &HeaterError{ID: id, Op: "Power", Err: err.Error()}
	}
	if err := heat.SetPower(pwr); err != nil {
		return &HeaterError{ID: id, Op: "Power.SetPower", Err: err.Error()}
	}
	return nil
}

func (h *HeaterHandler) StatusBy(id string) (HeaterConfig, error) {
	heat, err := h.by(id)
	if err != nil {
		return HeaterConfig{}, &HeaterError{ID: id, Op: "StatusBy", Err: err.Error()}
	}
	return HeaterConfig{
		ID:      id,
		Enabled: heat.Enabled(),
		Power:   heat.Power(),
	}, nil
}

func (h *HeaterHandler) Status() []HeaterConfig {
	status := make([]HeaterConfig, len(h.heaters))
	pos := 0
	for id, heat := range h.heaters {
		status[pos] = HeaterConfig{
			ID:      id,
			Enabled: heat.Enabled(),
			Power:   heat.Power(),
		}
		pos++
	}
	return status
}

func (h *HeaterHandler) by(id string) (Heater, error) {
	maybeHeater, ok := h.heaters[id]
	if !ok {
		return nil, ErrHeaterDoesntExist
	}
	return maybeHeater, nil
}

func (h *HeaterHandler) Open() {
}

func (h *HeaterHandler) Close() {
}
