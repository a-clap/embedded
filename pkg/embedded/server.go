/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"github.com/gin-gonic/gin"
)

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

var (
	ErrNotImplemented = errors.New("not implemented")
)

// routes configures default handlers for paths above
func (h *Handler) routes() {
	h.GET(RoutesGetHeaters, h.getHeaters())
	h.PUT(RoutesConfigHeater, h.configHeater())

	h.GET(RoutesGetOnewireSensors, h.getOnewireSensors())
	h.GET(RoutesGetOnewireTemperatures, h.getOnewireTemperatures())
	h.PUT(RoutesConfigOnewireSensor, h.configOnewireSensor())

	h.GET(RoutesGetPT100Sensors, h.getPTSensors())
	h.GET(RoutesGetPT100Temperatures, h.getPTTemperatures())
	h.PUT(RoutesConfigPT100Sensor, h.configPTSensor())

	h.GET(RoutesGetGPIOs, h.getGPIOS())
	h.PUT(RoutesConfigGPIO, h.configGPIO())
}

// common respond for whole rest API
func (*Handler) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
