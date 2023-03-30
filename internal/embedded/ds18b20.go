/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"net/http"
	"time"

	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/gin-gonic/gin"
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

type DSSensor interface {
	Name() (bus string, id string)
	Poll() (err error)
	Temperature() (actual, average float64, err error)
	GetReadings() []ds18b20.Readings
	Average() float64
	Configure(config ds18b20.SensorConfig) error
	GetConfig() ds18b20.SensorConfig
	Close() error
}

type DSSensorConfig struct {
	Bus     string `json:"bus"`
	Enabled bool   `json:"enabled"`
	ds18b20.SensorConfig
}

type DSTemperature struct {
	Bus      string             `json:"bus"`
	Readings []ds18b20.Readings `json:"readings"`
}

type dsSensor struct {
	DSSensor
	cfg DSSensorConfig
}

type DSHandler struct {
	sensors map[string]*dsSensor
}

func (h *Handler) configOnewireSensor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.DS.sensors == nil {
			e := &Error{
				Title:     "Failed to Configure",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		cfg := DSSensorConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			e := &Error{
				Title:     "Failed to bind DSSensorConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusBadRequest, e)
			return
		}

		cfg, err := h.DS.SetConfig(cfg)
		if err != nil {
			e := &Error{
				Title:     "Failed to SetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}

		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getOnewireTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.DS.sensors == nil {
			e := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetOnewireTemperatures,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}
		temperatures := h.DS.GetTemperatures()
		h.respond(ctx, http.StatusOK, temperatures)
	}
}

func (h *Handler) getOnewireSensors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sensors []DSSensorConfig
		if h.DS.sensors != nil {
			sensors = h.DS.GetSensors()
		}
		if len(sensors) == 0 {
			e := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetOnewireSensors,
				Timestamp: time.Now(),
			}
			h.respond(ctx, http.StatusInternalServerError, e)
			return
		}
		h.respond(ctx, http.StatusOK, sensors)
	}
}

func (d *DSHandler) GetTemperatures() []DSTemperature {
	sensors := make([]DSTemperature, 0, len(d.sensors))

	for _, s := range d.sensors {
		bus, _ := s.Name()
		tmp := DSTemperature{
			Bus:      bus,
			Readings: s.GetReadings(),
		}
		sensors = append(sensors, tmp)
	}

	return sensors
}
func (d *DSHandler) Temperature(cfg ds18b20.SensorConfig) (float64, float64, error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return 0, 0, &DSError{ID: cfg.ID, Op: "Temperature", Err: err.Error()}
	}
	actual, average, err := ds.Temperature()
	if err != nil {
		err = &DSError{ID: cfg.ID, Op: "Temperature", Err: err.Error()}
	}
	return actual, average, err
}

func (d *DSHandler) SetConfig(cfg DSSensorConfig) (newConfig DSSensorConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		err = &DSError{ID: cfg.ID, Op: "SetConfig.sensoryBy", Err: err.Error()}
		return
	}

	if err = ds.Configure(cfg.SensorConfig); err != nil {
		err = &DSError{ID: cfg.ID, Op: "SetConfig.Configure", Err: err.Error()}
		return
	}

	if cfg.Enabled != ds.cfg.Enabled {
		if cfg.Enabled {
			if err = ds.Poll(); err != nil {
				err = &DSError{ID: cfg.ID, Op: "SetConfig.Poll", Err: err.Error()}
				return
			}
		} else {
			if err = ds.Close(); err != nil {
				err = &DSError{ID: cfg.ID, Op: "SetConfig.Close", Err: err.Error()}
				return
			}
		}
	}
	ds.cfg.Enabled = cfg.Enabled

	return d.GetConfig(cfg.ID)
}

func (d *DSHandler) GetConfig(id string) (DSSensorConfig, error) {
	s, err := d.sensorBy(id)
	if err != nil {
		return DSSensorConfig{}, &DSError{ID: id, Op: "GetConfig", Err: err.Error()}
	}
	s.cfg.SensorConfig = s.GetConfig()
	return s.cfg, nil
}

func (d *DSHandler) sensorBy(id string) (*dsSensor, error) {
	if s, ok := d.sensors[id]; ok {
		return s, nil
	}
	return nil, ErrNoSuchID
}

func (d *DSHandler) GetSensors() []DSSensorConfig {
	onewireSensors := make([]DSSensorConfig, 0, len(d.sensors))
	for _, v := range d.sensors {
		onewireSensors = append(onewireSensors, v.cfg)
	}
	return onewireSensors
}

func (d *DSHandler) Open() {
}

func (d *DSHandler) Close() []error {
	var errs []error
	for name, sensor := range d.sensors {
		if err := sensor.Close(); err != nil {
			err = &DSError{ID: name, Op: "Close", Err: err.Error()}
			errs = append(errs, err)
		}
	}
	return errs
}
