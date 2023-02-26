/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
		if len(h.Heaters.heaters) == 0 {
			e := &Error{
				Title:     "Failed to Config",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		cfg := HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			e := &Error{
				Title:     "Failed to bind HeaterConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		if err := h.Heaters.Config(cfg); err != nil {
			e := &Error{
				Title:     "Failed to Config",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		s, _ := h.Heaters.StatusBy(cfg.ID)
		h.respond(ctx, http.StatusOK, s)
	}
}

func (h *Handler) getHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var heaters []HeaterConfig
		if len(h.Heaters.heaters) != 0 {
			heaters = h.Heaters.Status()
		}
		if len(heaters) == 0 {
			e := &Error{
				Title:     "Failed to Config",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
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
		return nil, ErrNoSuchID
	}
	return maybeHeater, nil
}

func (h *HeaterHandler) Open() {
}

func (h *HeaterHandler) Close() {
}
