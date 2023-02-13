/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	ErrNoSuchSensor = errors.New("specified dsSensor doesnt' exist")
)

type DSSensor interface {
	Name() (bus string, id string)
	Poll() (err error)
	Temperature() (actual, average float32, err error)
	GetReadings() []ds18b20.Readings
	Average() float32
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
		cfg := DSSensorConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		cfg, err := h.DS.SetConfig(cfg)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getOnewireTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ds := h.DS.GetTemperatures()
		if len(ds) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, ds)
	}
}

func (h *Handler) getOnewireSensors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ds := h.DS.GetSensors()
		if len(ds) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, ds)
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
func (d *DSHandler) Temperature(cfg ds18b20.SensorConfig) (float32, float32, error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return 0, 0, err
	}

	return ds.Temperature()
}

func (d *DSHandler) SetConfig(cfg DSSensorConfig) (newConfig DSSensorConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = ds.Configure(cfg.SensorConfig); err != nil {
		return
	}

	if cfg.Enabled != ds.cfg.Enabled {
		if cfg.Enabled {
			if err = ds.Poll(); err != nil {
				return
			}
		} else {
			if err = ds.Close(); err != nil {
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
		return DSSensorConfig{}, err
	}
	s.cfg.SensorConfig = s.GetConfig()
	return s.cfg, nil
}

func (d *DSHandler) sensorBy(id string) (*dsSensor, error) {
	if s, ok := d.sensors[id]; ok {
		return s, nil
	}
	return nil, ErrNoSuchSensor
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

func (d *DSHandler) Close() {
	for name, sensor := range d.sensors {
		if err := sensor.Close(); err != nil {
			log.Error("close ", name, " :", err)
		}
	}
}
