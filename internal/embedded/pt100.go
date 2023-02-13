/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PTSensor interface {
	ID() string
	Poll() (err error)
	Configure(config max31865.SensorConfig) error
	GetConfig() max31865.SensorConfig
	Average() float32
	Temperature() (actual float32, average float32, err error)
	GetReadings() []max31865.Readings
	Close() error
}

type PTSensorConfig struct {
	Enabled bool `json:"enabled"`
	max31865.SensorConfig
}

type ptSensor struct {
	PTSensor
	PTSensorConfig
}

type PTTemperature struct {
	Readings []max31865.Readings
}

// PTHandler is responsible for handling models.PTSensors
type PTHandler struct {
	sensors map[string]*ptSensor
}

// configPTSensor is middleware for configuring specified by ID PTSensor
func (h *Handler) configPTSensor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := PTSensorConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		cfg, err := h.PT.SetConfig(cfg)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getPTTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ds := h.PT.GetTemperatures()
		if len(ds) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, ds)
	}
}

func (h *Handler) getPTSensors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pt := h.PT.GetSensors()
		if len(pt) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, pt)
	}
}

func (p *PTHandler) GetTemperatures() []PTTemperature {
	temps := make([]PTTemperature, 0, len(p.sensors))
	for _, pt := range p.sensors {
		tmp := PTTemperature{Readings: pt.GetReadings()}
		temps = append(temps, tmp)
	}
	return temps
}

func (p *PTHandler) GetSensors() []PTSensorConfig {
	sensors := make([]PTSensorConfig, 0, len(p.sensors))
	for _, pt := range p.sensors {
		sensors = append(sensors, pt.PTSensorConfig)
	}
	return sensors
}

func (p *PTHandler) SetConfig(cfg PTSensorConfig) (newCfg PTSensorConfig, err error) {
	sensor, err := p.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = sensor.Configure(cfg.SensorConfig); err != nil {
		return
	}

	if cfg.Enabled != sensor.Enabled {
		if cfg.Enabled {
			if err = sensor.Poll(); err != nil {
				return
			}
		} else {
			if err = sensor.Close(); err != nil {
				return
			}
		}
	}
	sensor.Enabled = cfg.Enabled

	return p.GetConfig(cfg.ID)
}

func (p *PTHandler) GetConfig(id string) (PTSensorConfig, error) {
	sensor, err := p.sensorBy(id)
	if err != nil {
		return PTSensorConfig{}, err
	}
	sensor.SensorConfig = sensor.GetConfig()
	return sensor.PTSensorConfig, nil
}

func (p *PTHandler) sensorBy(id string) (*ptSensor, error) {
	maybeSensor, ok := p.sensors[id]
	if !ok {
		return nil, ErrNoSuchSensor
	}
	return maybeSensor, nil
}

func (p *PTHandler) Open() {
}
func (p *PTHandler) Close() {
	for name, sensor := range p.sensors {
		if sensor.Enabled {
			sensor.Enabled = false
			if err := sensor.Close(); err != nil {
				log.Error("close failed on sensor: ", name, ", with error ", err)
			}
		}
	}
}
