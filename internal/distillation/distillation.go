/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	HeatersHandler *HeatersHandler
	DSHandler      *DSHandler
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{Engine: gin.Default()}
	// Options
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	h.routes()

	return h, nil
}

const (
	RoutesConfigHeater        = "/api/heater/all"
	RoutesGetAllHeaters       = "/api/heater/all"
	RoutesGetEnabledHeaters   = "/api/heater/phase"
	RoutesConfigEnabledHeater = "/api/heater/phase"
	RoutesGetDS               = "/api/ds"
	RoutesConfigureDS         = "/api/ds"
	RoutesGetDSTemperatures   = "/api/ds/temperatures"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// routes configures default handlers for paths above
func (h *Handler) routes() {
	h.GET(RoutesGetAllHeaters, h.getAllHeaters())
	h.GET(RoutesGetEnabledHeaters, h.getEnabledHeaters())
	h.PUT(RoutesConfigHeater, h.configHeater())
	h.PUT(RoutesConfigEnabledHeater, h.configEnabledHeater())

	h.GET(RoutesGetDS, h.getDS())
	h.GET(RoutesGetDSTemperatures, h.getDSTemperatures())
	h.PUT(RoutesConfigureDS, h.configureDS())
}

// common respond for whole rest API
func (*Handler) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
