package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	RoutesGetSensor = "/api/sensors"
)

func (s *Server) routes() {
	s.GET(RoutesGetSensor, s.getSensors())
}

func (s *Server) getSensors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.GetSensors == nil {
			s.write(c, http.StatusInternalServerError, gin.H{"failed": "really"})
			return
		}
	}
}
