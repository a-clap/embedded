/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Sensor struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Temperature float32 `json:"temperature"`
}

type SensorHandler interface {
	Sensors() ([]Sensor, error)
}

func (s *Server) getSensors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.SensorHandler == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		sensors, err := s.SensorHandler.Sensors()
		if err != nil {
			s.write(c, http.StatusInternalServerError, ErrNotFound)
			return
		}
		s.write(c, http.StatusOK, sensors)

	}
}
