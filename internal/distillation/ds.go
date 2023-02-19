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
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/gin-gonic/gin"
)

var (
	ErrNoSuchDS      = errors.New("doesn't exist")
	ErrNoTemps       = errors.New("temperature buffer is empty")
	ErrUnexpectedID  = errors.New("unexpected ID")
	ErrNoDSInterface = errors.New("no ds interface")
)

type DSError struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (d *DSError) Error() string {
	if d.Err == "" {
		return "<nil>"
	}
	s := d.Op
	if d.ID != "" {
		s += ":" + d.ID
	}
	s += ": " + d.Err
	return s
}

// DS access to on-board DS18B20 sensors
type DS interface {
	Get() ([]embedded.DSSensorConfig, error)
	Set(s embedded.DSSensorConfig) error
	Temperatures() ([]embedded.DSTemperature, error)
}

// DSConfig simple wrapper for sensor configuration
type DSConfig struct {
	embedded.DSSensorConfig
	temps embedded.DSTemperature
}

// DSHandler main struct used to handle number of DS sensors
type DSHandler struct {
	DS           DS
	sensors      map[string]*DSConfig
	pollInterval time.Duration
}

// DSTemperature - json returned from rest API
type DSTemperature struct {
	ID          string  `json:"ID"`
	Temperature float32 `json:"temperature"`
}

func (h *Handler) getDS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sensors := h.DSHandler.GetSensors()
		if len(sensors) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, sensors)
	}
}

func (h *Handler) getDSTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		temperatures := h.DSHandler.Temperatures()
		if len(temperatures) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, temperatures)
	}
}

func (h *Handler) configureDS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := DSConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		if err := h.DSHandler.ConfigureSensor(cfg); err != nil {
			if dsErr, ok := err.(*DSError); ok {
				h.respond(ctx, http.StatusInternalServerError, dsErr)
			} else {
				h.respond(ctx, http.StatusInternalServerError, err)
			}
			return
		}
		cfg, err := h.DSHandler.GetConfig(cfg.ID)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		h.respond(ctx, http.StatusOK, cfg)
	}
}

// NewDSHandler creates new DSHandler with provided DS interface
func NewDSHandler(ds DS) (*DSHandler, error) {
	d := &DSHandler{
		DS:           ds,
		sensors:      make(map[string]*DSConfig),
		pollInterval: 1 * time.Second,
	}
	if err := d.init(); err != nil {
		return nil, err
	}
	return d, nil
}

// Update updates temperatures in sensors
func (d *DSHandler) Update() (errs []error) {
	temps, err := d.DS.Temperatures()
	if err != nil {
		errs = append(errs, &DSError{Op: "Update.Temperatures", Err: err.Error()})
		return
	}
	for _, temp := range temps {
		if len(temp.Readings) == 0 {
			continue
		}
		for _, single := range temp.Readings {
			id := single.ID
			s, ok := d.sensors[id]
			if !ok {
				errs = append(errs, &DSError{ID: id, Op: "Update.Temperatures", Err: ErrUnexpectedID.Error()})
				continue
			}
			s.temps.Readings = append(s.temps.Readings, single)
		}
	}
	return
}

// History returns historical temperatures, but it also CLEARS all history data but last
func (d *DSHandler) History() []embedded.DSTemperature {
	t := make([]embedded.DSTemperature, 0, len(d.sensors))
	for _, v := range d.sensors {
		length := len(v.temps.Readings)
		if length < 2 {
			continue
		}
		var data []ds18b20.Readings
		data, v.temps.Readings = v.temps.Readings[0:length-1], v.temps.Readings[length-1:]

		t = append(t, embedded.DSTemperature{
			Bus:      v.Bus,
			Readings: data,
		})
	}
	return t
}

// Temperatures returns last read temperature for all sensors
func (d *DSHandler) Temperatures() []DSTemperature {
	t := make([]DSTemperature, 0, len(d.sensors))
	for id := range d.sensors {
		temp, _ := d.Temperature(id)
		t = append(t, DSTemperature{
			ID:          id,
			Temperature: temp,
		})
	}
	return t
}

// Temperature returns last read temperature
func (d *DSHandler) Temperature(id string) (float32, error) {
	ds, ok := d.sensors[id]
	if !ok {
		return 0.0, &DSError{ID: id, Op: "Temperature", Err: ErrNoSuchDS.Error()}
	}

	size := len(ds.temps.Readings)
	if size == 0 {
		return 0.0, &DSError{ID: id, Op: "Temperature", Err: ErrNoTemps.Error()}
	}
	// Return last temperature
	return ds.temps.Readings[size-1].Average, nil
}

func (d *DSHandler) ConfigureSensor(cfg DSConfig) error {
	ds, ok := d.sensors[cfg.ID]
	if !ok {
		return &DSError{ID: cfg.ID, Op: "ConfigureSensor", Err: ErrNoSuchDS.Error()}
	}
	if err := d.DS.Set(cfg.DSSensorConfig); err != nil {
		return &DSError{ID: cfg.ID, Op: "ConfigureSensor.Set", Err: err.Error()}
	}
	ds.DSSensorConfig = cfg.DSSensorConfig
	return nil
}

func (d *DSHandler) GetConfig(id string) (DSConfig, error) {
	ds, ok := d.sensors[id]
	if !ok {
		return DSConfig{}, &DSError{ID: id, Op: "GetConfig", Err: ErrNoSuchDS.Error()}
	}
	return *ds, nil
}

func (d *DSHandler) GetSensors() []DSConfig {
	s := make([]DSConfig, 0, len(d.sensors))
	for _, elem := range d.sensors {
		s = append(s, *elem)
	}
	return s
}

func (d *DSHandler) init() error {
	if d.DS == nil {
		return &DSError{Op: "init", Err: ErrNoDSInterface.Error()}
	}
	sensors, err := d.DS.Get()
	if err != nil {
		return &DSError{Op: "init.Get", Err: err.Error()}
	}

	for _, ds := range sensors {
		id := ds.ID
		cfg := &DSConfig{
			DSSensorConfig: ds,
			temps:          embedded.DSTemperature{},
		}
		d.sensors[id] = cfg
		// TODO: Should we configure them on startup?
	}
	return nil
}
