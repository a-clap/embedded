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

type GetSensors interface {
	Sensors() ([]Sensor, error)
}

func (s *Server) getSensors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.GetSensors == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		sensors, err := s.GetSensors.Sensors()
		if err != nil {
			s.write(c, http.StatusInternalServerError, ErrNotFound)
			return
		}
		s.write(c, http.StatusOK, sensors)

	}
}
