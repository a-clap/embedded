package rest

const (
	RoutesGetSensor = "/api/sensors"
)

func (s *Server) routes() {
	s.GET(RoutesGetSensor, s.getSensors())
}
