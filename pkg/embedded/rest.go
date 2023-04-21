/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
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

// routes configures default handlers for paths above
func (e *Embedded) routes() {
	e.GET(RoutesGetHeaters, e.getHeaters())
	e.PUT(RoutesConfigHeater, e.configHeater())
	
	e.GET(RoutesGetOnewireSensors, e.getOnewireSensors())
	e.GET(RoutesGetOnewireTemperatures, e.getOnewireTemperatures())
	e.PUT(RoutesConfigOnewireSensor, e.configOnewireSensor())
	
	e.GET(RoutesGetPT100Sensors, e.getPTSensors())
	e.GET(RoutesGetPT100Temperatures, e.getPTTemperatures())
	e.PUT(RoutesConfigPT100Sensor, e.configPTSensor())
	
	e.GET(RoutesGetGPIOs, e.getGPIOS())
	e.PUT(RoutesConfigGPIO, e.configGPIO())
}

// common respond for whole rest API
func (*Embedded) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
