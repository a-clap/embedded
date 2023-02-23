/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	RoutesEnableHeater      = "/api/heater"
	RoutesGetAllHeaters     = "/api/heater"
	RoutesGetEnabledHeaters = "/api/heater/enabled"
	RoutesConfigureHeater   = "/api/heater/enabled"
	RoutesGetDS             = "/api/onewire"
	RoutesConfigureDS       = "/api/onewire"
	RoutesGetDSTemperatures = "/api/onewire/temperatures"
	RoutesGetPT             = "/api/pt100"
	RoutesConfigurePT       = "/api/pt100"
	RoutesGetPTTemperatures = "/api/pt100/temperatures"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// routes configures default handlers for paths above
func (h *Handler) routes() {
	h.GET(RoutesGetAllHeaters, h.getAllHeaters())
	h.GET(RoutesGetEnabledHeaters, h.getEnabledHeaters())
	h.PUT(RoutesEnableHeater, h.enableHeater())
	h.PUT(RoutesConfigureHeater, h.configEnabledHeater())

	h.GET(RoutesGetDS, h.getDS())
	h.GET(RoutesGetDSTemperatures, h.getDSTemperatures())
	h.PUT(RoutesConfigureDS, h.configureDS())

	h.GET(RoutesGetPT, h.getPT())
	h.GET(RoutesGetPTTemperatures, h.getPTTemperatures())
	h.PUT(RoutesConfigurePT, h.configurePT())
}

// common respond for whole rest API
func (*Handler) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
