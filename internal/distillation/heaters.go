/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"errors"
	"net/http"
	"time"

	"github.com/a-clap/iot/internal/embedded"
	"github.com/gin-gonic/gin"
)

var (
	ErrNoSuchHeater = errors.New("no such heater with id: ")
)

type Heaters interface {
	Get() ([]embedded.HeaterConfig, error)
	Set(heater embedded.HeaterConfig) error
}

type HeaterConfigGlobal struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type HeaterConfig struct {
	global HeaterConfigGlobal
	embedded.HeaterConfig
}

type HeatersHandler struct {
	Heaters Heaters
	heaters map[string]*HeaterConfig
}

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

func (h *Handler) configEnabledHeater() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := embedded.HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			e := &Error{
				Title:     "Failed to bind HeaterConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigEnabledHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		if err := h.HeatersHandler.Configure(cfg); err != nil {
			e := &Error{
				Title:     "Failed to Configure",
				Detail:    err.Error(),
				Instance:  RoutesConfigEnabledHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		newCfg, err := h.HeatersHandler.Config(cfg.ID)
		if err != nil {
			e := &Error{
				Title:     "Failed to Config",
				Detail:    err.Error(),
				Instance:  RoutesConfigEnabledHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		h.respond(ctx, http.StatusOK, newCfg)
	}
}
func (h *Handler) configHeater() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := HeaterConfigGlobal{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			e := &Error{
				Title:     "Failed to bind HeaterConfigGlobal",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}
		if err := h.HeatersHandler.ConfigureGlobal(cfg); err != nil {
			e := &Error{
				Title:     "Failed to ConfigureGlobal",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		newCfg, err := h.HeatersHandler.ConfigGlobal(cfg.ID)
		if err != nil {
			e := &Error{
				Title:     "Failed to ConfigGlobal",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		h.respond(ctx, http.StatusOK, newCfg)
	}
}

func (h *Handler) getAllHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		heaters := h.HeatersHandler.ConfigsGlobal()
		if len(heaters) == 0 {
			e := &Error{
				Title:     "Failed to ConfigsGlobal",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetAllHeaters,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}
		h.respond(ctx, http.StatusOK, heaters)
	}
}

func (h *Handler) getEnabledHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		heaters := h.HeatersHandler.Configs()
		if len(heaters) == 0 {
			e := &Error{
				Title:     "Failed to ConfigsGlobal",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetEnabledHeaters,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}
		h.respond(ctx, http.StatusOK, heaters)
	}
}

func NewHandlerHeaters(heaters Heaters) (*HeatersHandler, error) {
	h := &HeatersHandler{
		Heaters: heaters,
		heaters: make(map[string]*HeaterConfig),
	}
	if err := h.init(); err != nil {
		return nil, &HeaterError{ID: "", Op: "NewHandlerHeaters.init", Err: err.Error()}
	}
	return h, nil
}

func (h *HeatersHandler) init() error {
	if h.Heaters == nil {
		return errors.New("no Heater interface")
	}
	heaters, err := h.Heaters.Get()
	if err != nil {
		return err
	}

	for _, heater := range heaters {
		id := heater.ID
		cfg := HeaterConfig{
			global: HeaterConfigGlobal{
				ID:      id,
				Enabled: false,
			},
			HeaterConfig: embedded.HeaterConfig{
				ID:      id,
				Enabled: false,
				Power:   0,
			},
		}

		h.heaters[id] = &cfg
		if err = h.Configure(cfg.HeaterConfig); err != nil {
			return err
		}

	}
	return nil
}

func (h *HeatersHandler) ConfigsGlobal() []HeaterConfigGlobal {
	heaters := make([]HeaterConfigGlobal, 0, len(h.heaters))
	for _, v := range h.heaters {
		heaters = append(heaters, v.global)
	}
	return heaters
}

func (h *HeatersHandler) ConfigureGlobal(cfg HeaterConfigGlobal) error {
	maybeHeater, ok := h.heaters[cfg.ID]
	if !ok {
		return &HeaterError{ID: cfg.ID, Op: "ConfigureGlobal", Err: ErrNoSuchHeater.Error()}
	}

	if maybeHeater.global.Enabled != cfg.Enabled {
		maybeHeater.global.Enabled = cfg.Enabled
		// Do we need to disable heater?
		if !maybeHeater.global.Enabled && maybeHeater.HeaterConfig.Enabled {
			maybeHeater.HeaterConfig.Enabled = false
			if err := h.Configure(maybeHeater.HeaterConfig); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HeatersHandler) Configs() []embedded.HeaterConfig {
	heaters := make([]embedded.HeaterConfig, 0, len(h.heaters))
	for _, v := range h.heaters {
		if v.global.Enabled {
			heaters = append(heaters, v.HeaterConfig)
		}
	}
	return heaters
}

func (h *HeatersHandler) Config(id string) (embedded.HeaterConfig, error) {
	cfg, ok := h.heaters[id]
	if !ok {
		return embedded.HeaterConfig{}, &HeaterError{ID: id, Op: "Config", Err: ErrNoSuchHeater.Error()}
	}
	return cfg.HeaterConfig, nil
}

func (h *HeatersHandler) ConfigGlobal(id string) (HeaterConfigGlobal, error) {
	cfg, ok := h.heaters[id]
	if !ok {
		return HeaterConfigGlobal{}, &HeaterError{ID: id, Op: "Config", Err: ErrNoSuchHeater.Error()}
	}
	return cfg.global, nil
}

func (h *HeatersHandler) Configure(cfg embedded.HeaterConfig) error {
	maybeHeater, ok := h.heaters[cfg.ID]
	if !ok {
		return &HeaterError{ID: cfg.ID, Op: "ConfigureGlobal", Err: ErrNoSuchHeater.Error()}
	}
	// Global has to be set
	maybeHeater.Enabled = maybeHeater.global.Enabled && cfg.Enabled
	maybeHeater.Power = cfg.Power
	err := h.Heaters.Set(maybeHeater.HeaterConfig)
	return err
}
