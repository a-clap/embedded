/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package rest

const (
	RoutesGetSensor   = "/api/sensors"
	RoutesGetWifiAps  = "/api/wifi/networks"
	RoutesConnectWifi = "/api/wifi"
)

func (s *Server) routes() {
	s.GET(RoutesGetSensor, s.getSensors())
	s.GET(RoutesGetWifiAps, s.getWifiAps())
	s.POST(RoutesConnectWifi, s.connectToAP())
}
