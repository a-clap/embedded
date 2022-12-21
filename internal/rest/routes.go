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
