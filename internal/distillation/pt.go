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
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/gin-gonic/gin"
)

var (
	ErrNoPTInterface = errors.New("no pt interface")
)

type PTError struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *PTError) Error() string {
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

// PT access to on-board PT100 sensors
type PT interface {
	Get() ([]embedded.PTSensorConfig, error)
	Set(s embedded.PTSensorConfig) error
	Temperatures() ([]embedded.PTTemperature, error)
}

// PTConfig simple wrapper for sensor configuration
type PTConfig struct {
	embedded.PTSensorConfig
	temps embedded.PTTemperature
}

// PTHandler main struct used to handle number of DS sensors
type PTHandler struct {
	PT           PT
	sensors      map[string]*PTConfig
	pollInterval time.Duration
}

// PTTemperature - json returned from rest API
type PTTemperature struct {
	ID          string  `json:"ID"`
	Temperature float32 `json:"temperature"`
}

func (h *Handler) getPT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensors := h.PTHandler.GetSensors()
		if len(sensors) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, sensors)
	}
}

func (h *Handler) getPTTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		temperatures := h.PTHandler.Temperatures()
		if len(temperatures) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, temperatures)
	}
}

func (h *Handler) configurePT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := PTConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		if err := h.PTHandler.ConfigureSensor(cfg); err != nil {
			if dsErr, ok := err.(*PTError); ok {
				h.respond(ctx, http.StatusInternalServerError, dsErr)
			} else {
				h.respond(ctx, http.StatusInternalServerError, err)
			}
			return
		}
		cfg, err := h.PTHandler.GetConfig(cfg.ID)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		h.respond(ctx, http.StatusOK, cfg)
	}
}

// NewPTHandler creates new PTHandler with provided PT interface
func NewPTHandler(pt PT) (*PTHandler, error) {
	d := &PTHandler{
		PT:      pt,
		sensors: make(map[string]*PTConfig),
	}
	if err := d.init(); err != nil {
		return nil, err
	}
	return d, nil
}

// Update updates temperatures in sensors
func (p *PTHandler) Update() (errs []error) {
	temps, err := p.PT.Temperatures()
	if err != nil {
		errs = append(errs, &PTError{Op: "Update.Temperatures", Err: err.Error()})
		return
	}
	for _, temp := range temps {
		if len(temp.Readings) == 0 {
			continue
		}
		for _, single := range temp.Readings {
			id := single.ID
			s, ok := p.sensors[id]
			if !ok {
				errs = append(errs, &PTError{ID: id, Op: "Update.Temperatures", Err: ErrUnexpectedID.Error()})
				continue
			}
			s.temps.Readings = append(s.temps.Readings, single)
		}
	}
	return
}

// History returns historical temperatures, but it also CLEARS all history data but last
func (p *PTHandler) History() []embedded.PTTemperature {
	t := make([]embedded.PTTemperature, 0, len(p.sensors))
	for _, v := range p.sensors {
		length := len(v.temps.Readings)
		if length < 2 {
			continue
		}
		var data []max31865.Readings
		data, v.temps.Readings = v.temps.Readings[0:length-1], v.temps.Readings[length-1:]

		t = append(t, embedded.PTTemperature{
			Readings: data,
		})
	}
	return t
}

// Temperatures returns last read temperature for all sensors
func (p *PTHandler) Temperatures() []PTTemperature {
	t := make([]PTTemperature, 0, len(p.sensors))
	for id := range p.sensors {
		temp, _ := p.Temperature(id)
		t = append(t, PTTemperature{
			ID:          id,
			Temperature: temp,
		})
	}
	return t
}

// Temperature returns last read temperature
func (p *PTHandler) Temperature(id string) (float32, error) {
	pt, ok := p.sensors[id]
	if !ok {
		return 0.0, &PTError{ID: id, Op: "Temperature", Err: ErrNoSuchID.Error()}
	}

	size := len(pt.temps.Readings)
	if size == 0 {
		return 0.0, &PTError{ID: id, Op: "Temperature", Err: ErrNoTemps.Error()}
	}
	// Return last temperature
	return pt.temps.Readings[size-1].Average, nil
}

func (p *PTHandler) ConfigureSensor(cfg PTConfig) error {
	pt, ok := p.sensors[cfg.ID]
	if !ok {
		return &PTError{ID: cfg.ID, Op: "ConfigureSensor", Err: ErrNoSuchID.Error()}
	}
	if err := p.PT.Set(cfg.PTSensorConfig); err != nil {
		return &PTError{ID: cfg.ID, Op: "ConfigureSensor.Set", Err: err.Error()}
	}
	pt.PTSensorConfig = cfg.PTSensorConfig
	return nil
}

func (p *PTHandler) GetConfig(id string) (PTConfig, error) {
	pt, ok := p.sensors[id]
	if !ok {
		return PTConfig{}, &PTError{ID: id, Op: "GetConfig", Err: ErrNoSuchID.Error()}
	}
	return *pt, nil
}

func (p *PTHandler) GetSensors() []PTConfig {
	s := make([]PTConfig, 0, len(p.sensors))
	for _, elem := range p.sensors {
		s = append(s, *elem)
	}
	return s
}

func (p *PTHandler) init() error {
	if p.PT == nil {
		return &PTError{Op: "init", Err: ErrNoPTInterface.Error()}
	}
	sensors, err := p.PT.Get()
	if err != nil {
		return &PTError{Op: "init.Get", Err: err.Error()}
	}

	for _, pt := range sensors {
		id := pt.ID
		cfg := &PTConfig{
			PTSensorConfig: pt,
			temps:          embedded.PTTemperature{},
		}
		p.sensors[id] = cfg
		// TODO: Should we configure them on startup?
	}
	return nil
}
