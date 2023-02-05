/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PTHandler struct {
	handlers []models.PTSensor
	sensors  map[string]models.PTSensor
}

func (h *Handler) configPTSensor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hwid := ctx.Param("hardware_id")
		if _, err := h.PT.GetConfig(hwid); err != nil {
			h.respond(ctx, http.StatusNotFound, err)
			return
		}

		cfg := models.PTConfig{}
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

func (p *PTHandler) GetTemperatures() []models.Temperature {
	temps := make([]models.Temperature, len(p.handlers))
	for i, pt := range p.handlers {
		temps[i] = pt.Temperature()
	}
	return temps
}

func (p *PTHandler) GetSensors() []models.PTConfig {
	sensors := make([]models.PTConfig, len(p.handlers))
	for i, pt := range p.handlers {
		sensors[i] = pt.Config()
	}
	return sensors
}

func (p *PTHandler) SetConfig(cfg models.PTConfig) (newCfg models.PTConfig, err error) {
	sensor, err := p.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = sensor.SetConfig(cfg); err != nil {
		return
	}

	return sensor.Config(), nil
}

func (p *PTHandler) GetConfig(hwid string) (models.PTConfig, error) {
	sensor, err := p.sensorBy(hwid)
	if err != nil {
		return models.PTConfig{}, err
	}
	return sensor.Config(), nil
}

func (p *PTHandler) sensorBy(hwid string) (models.PTSensor, error) {
	maybeSensor, ok := p.sensors[hwid]
	if !ok {
		return nil, ErrNoSuchSensor
	}
	return maybeSensor, nil
}

func (p *PTHandler) Open() {
	if p.handlers == nil {
		return
	}
	p.sensors = make(map[string]models.PTSensor)

	for _, elem := range p.handlers {
		p.sensors[elem.ID()] = elem
	}
}
func (p *PTHandler) Close() {
	for name, sensor := range p.sensors {
		cfg := sensor.Config()
		if cfg.Enabled {
			cfg.Enabled = false
			err := sensor.SetConfig(cfg)
			if err != nil {
				log.Debug("SetConfig failed on sensor: ", name)
			}
		}
	}
}
