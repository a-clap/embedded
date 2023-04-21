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

type restRouter struct {
	*gin.Engine
}

const (
	RoutesGetHeaters             = "/api/heater"
	RoutesConfigHeater           = "/api/heater"
	RoutesGetOnewireSensors      = "/api/onewire"
	RoutesGetOnewireTemperatures = "/api/onewire/temperatures"
	RoutesConfigOnewireSensor    = "/api/onewire"
	RoutesGetPT100Sensors        = "/api/pt100"
	RoutesGetPT100Temperatures   = "/api/pt100/temperatures"
	RoutesConfigPT100Sensor      = "/api/pt100"
	RoutesGetGPIOs               = "/api/gpio"
	RoutesConfigGPIO             = "/api/gpio"
)

func (r *restRouter) routes(e *Embedded) {
	r.GET(RoutesGetHeaters, r.getHeaters(e))
	r.PUT(RoutesConfigHeater, r.configHeater(e))
	
	r.GET(RoutesGetOnewireSensors, r.getOnewireSensors(e))
	r.GET(RoutesGetOnewireTemperatures, r.getOnewireTemperatures(e))
	r.PUT(RoutesConfigOnewireSensor, r.configOnewireSensor(e))
	
	r.GET(RoutesGetPT100Sensors, r.getPTSensors(e))
	r.GET(RoutesGetPT100Temperatures, r.getPTTemperatures(e))
	r.PUT(RoutesConfigPT100Sensor, r.configPTSensor(e))
	
	r.GET(RoutesGetGPIOs, r.getGPIOS(e))
	r.PUT(RoutesConfigGPIO, r.configGPIO(e))
}

// common respond for whole rest API
func (*restRouter) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}

func (r *restRouter) configHeater(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(e.Heaters.heaters) == 0 {
			err := &Error{
				Title:     "Failed to Config",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		cfg := HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			err := &Error{
				Title:     "Failed to bind HeaterConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusBadRequest, err)
			return
		}
		
		if err := e.Heaters.Config(cfg); err != nil {
			err := &Error{
				Title:     "Failed to Config",
				Detail:    err.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		s, _ := e.Heaters.StatusBy(cfg.ID)
		r.respond(ctx, http.StatusOK, s)
	}
}

func (r *restRouter) getHeaters(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var heaters []HeaterConfig
		if len(e.Heaters.heaters) != 0 {
			heaters = e.Heaters.Status()
		}
		if len(heaters) == 0 {
			err := &Error{
				Title:     "Failed to Config",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigHeater,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		r.respond(ctx, http.StatusOK, heaters)
	}
}

func (r *restRouter) configOnewireSensor(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.DS.sensors == nil {
			err := &Error{
				Title:     "Failed to Configure",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		cfg := DSSensorConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			err := &Error{
				Title:     "Failed to bind DSSensorConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusBadRequest, err)
			return
		}
		
		cfg, err := e.DS.SetConfig(cfg)
		if err != nil {
			err := &Error{
				Title:     "Failed to SetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigOnewireSensor,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		r.respond(ctx, http.StatusOK, cfg)
	}
}
func (r *restRouter) getOnewireTemperatures(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.DS.sensors == nil {
			err := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetOnewireTemperatures,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		temperatures := e.DS.GetTemperatures()
		r.respond(ctx, http.StatusOK, temperatures)
	}
}

func (r *restRouter) getOnewireSensors(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sensors []DSSensorConfig
		if e.DS.sensors != nil {
			sensors = e.DS.GetSensors()
		}
		if len(sensors) == 0 {
			err := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetOnewireSensors,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		r.respond(ctx, http.StatusOK, sensors)
	}
}

// configPTSensor is middleware for configuring specified by ID PTSensor
func (r *restRouter) configPTSensor(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.PT.sensors == nil {
			err := &Error{
				Title:     "Failed to Configure",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigPT100Sensor,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
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
			r.respond(ctx, http.StatusBadRequest, err)
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
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		r.respond(ctx, http.StatusOK, cfg)
	}
}
func (r *restRouter) getPTTemperatures(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if e.PT.sensors == nil {
			err := &Error{
				Title:     "Failed to GetTemperatures",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesGetPT100Temperatures,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		temperatures := e.PT.GetTemperatures()
		r.respond(ctx, http.StatusOK, temperatures)
	}
}

func (r *restRouter) getPTSensors(e *Embedded) gin.HandlerFunc {
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
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		r.respond(ctx, http.StatusOK, sensors)
	}
}

func (r *restRouter) configGPIO(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(e.GPIO.io) == 0 {
			err := &Error{
				Title:     "Failed to Config GPIO",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		cfg := GPIOConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			err := &Error{
				Title:     "Failed to bind GPIOConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusBadRequest, err)
			return
		}
		
		err := e.GPIO.SetConfig(cfg)
		if err != nil {
			err := &Error{
				Title:     "Failed to SetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		cfg, err = e.GPIO.GetConfig(cfg.ID)
		if err != nil {
			err := &Error{
				Title:     "Failed to GetConfig",
				Detail:    err.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		r.respond(ctx, http.StatusOK, cfg)
	}
}
func (r *restRouter) getGPIOS(e *Embedded) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(e.GPIO.io) == 0 {
			err := &Error{
				Title:     "Failed to GetGPIO",
				Detail:    ErrNotImplemented.Error(),
				Instance:  RoutesConfigGPIO,
				Timestamp: time.Now(),
			}
			r.respond(ctx, http.StatusInternalServerError, err)
			return
		}
		
		gpios, err := e.GPIO.GetConfigAll()
		if len(gpios) == 0 || err != nil {
			notImpl := GPIOError{ID: "", Op: "GetConfigAll", Err: ErrNotImplemented.Error()}
			r.respond(ctx, http.StatusInternalServerError, &notImpl)
			return
		}
		r.respond(ctx, http.StatusOK, gpios)
	}
}
