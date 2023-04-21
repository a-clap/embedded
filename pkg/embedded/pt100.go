/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"net/http"
	"time"
	
	"github.com/a-clap/embedded/pkg/max31865"
	"github.com/gin-gonic/gin"
)

type PTSensor interface {
	ID() string
	Poll() (err error)
	Configure(config max31865.SensorConfig) error
	GetConfig() max31865.SensorConfig
	Average() float64
	Temperature() (actual float64, average float64, err error)
	GetReadings() []max31865.Readings
	Close() error
}
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
func (e *Embedded) configPTSensor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.PT.sensors == nil {
			err := &Error{
				Title:     "Failed to Configure",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigPT100Sensor,
				Timestamp: time.Now(),
			}
			e.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		cfg := PTSensorConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			err := &Error{
				Title:     "Failed to bind PTSensorConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigPT100Sensor,
				Timestamp: time.Now(),
			}
			e.respond(ctx, http.StatusBadRequest, err)
			return
		}
		
		cfg, err := e.PT.SetConfig(cfg)
		if err != nil {
			err := &Error{
				Title:     "Failed to SetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigPT100Sensor,
				Timestamp: time.Now(),
			}
			e.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		e.respond(ctx, http.StatusOK, cfg)
	}
}
func (e *Embedded) getPTTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.PT.sensors == nil {
			err := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetPT100Temperatures,
				Timestamp: time.Now(),
			}
			e.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		temperatures := e.PT.GetTemperatures()
		e.respond(ctx, http.StatusOK, temperatures)
	}
}

func (e *Embedded) getPTSensors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sensors []PTSensorConfig
		if e.PT.sensors != nil {
			sensors = e.PT.GetSensors()
		}
		if len(sensors) == 0 {
			err := &Error{
				Title:     "Failed to GetSensors",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetPT100Temperatures,
				Timestamp: time.Now(),
			}
			e.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		e.respond(ctx, http.StatusOK, sensors)
	}
}

func (p *PTHandler) GetTemperatures() []PTTemperature {
	temps := make([]PTTemperature, 0, len(p.sensors))
	for _, pt := range p.sensors {
		if pt.Enabled {
			tmp := PTTemperature{Readings: pt.GetReadings()}
			temps = append(temps, tmp)
		}
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
		err = &PTError{ID: cfg.ID, Op: "SetConfig.sensorBy", Err: err.Error()}
		return
	}
	
	if err = sensor.Configure(cfg.SensorConfig); err != nil {
		err = &PTError{ID: cfg.ID, Op: "SetConfig.Configure", Err: err.Error()}
		return
	}
	
	if cfg.Enabled != sensor.Enabled {
		if cfg.Enabled {
			if err = sensor.Poll(); err != nil {
				err = &PTError{ID: cfg.ID, Op: "SetConfig.Poll", Err: err.Error()}
				return
			}
		} else {
			if err = sensor.Close(); err != nil {
				err = &PTError{ID: cfg.ID, Op: "SetConfig.Close", Err: err.Error()}
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
		return PTSensorConfig{}, &PTError{ID: id, Op: "GetConfig.sensorBy", Err: err.Error()}
	}
	sensor.SensorConfig = sensor.GetConfig()
	return sensor.PTSensorConfig, nil
}

func (p *PTHandler) sensorBy(id string) (*ptSensor, error) {
	maybeSensor, ok := p.sensors[id]
	if !ok {
		return nil, ErrNoSuchID
	}
	return maybeSensor, nil
}

func (p *PTHandler) Open() {
}

func (p *PTHandler) Close() []error {
	var errs []error
	for name, sensor := range p.sensors {
		if sensor.Enabled {
			sensor.Enabled = false
			if err := sensor.Close(); err != nil {
				err = &PTError{ID: name, Op: "Close", Err: err.Error()}
				errs = append(errs, err)
			}
		}
	}
	return errs
}
