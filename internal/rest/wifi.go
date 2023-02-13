/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WIFINetwork struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

type WIFIStatus struct {
	Connected bool   `json:"connected"`
	SSID      string `json:"ssid"`
}

type WIFIHandler interface {
	APs() ([]WIFINetwork, error)
	Connect(n WIFINetwork) error
	Disconnect() error
	Status() (WIFIStatus, error)
}

func (s *Server) getWifiAps() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.WIFIHandler == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		aps, err := s.WIFIHandler.APs()
		if err != nil {
			s.write(c, http.StatusInternalServerError, errorInterface(err))
			return
		}
		s.write(c, http.StatusOK, aps)
	}
}

func (s *Server) connectToAP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.WIFIHandler == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		ap := WIFINetwork{}
		if err := c.ShouldBind(&ap); err != nil {
			s.write(c, http.StatusBadRequest, ErrInterface)
			return
		}

		if err := s.Connect(ap); err != nil {
			s.write(c, http.StatusInternalServerError, errorInterface(err))
			return
		}

		s.write(c, http.StatusOK, nil)
	}
}
