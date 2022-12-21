package rest

import (
	"github.com/a-clap/iot/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WifiHandler interface {
	APs() ([]models.WifiNetwork, error)
	Connect(n models.WifiNetwork) error
	Disconnect() error
	Status() (models.WifiStatus, error)
}

func (s *Server) getWifiAps() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.WifiHandler == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		aps, err := s.WifiHandler.APs()
		if err != nil {
			s.write(c, http.StatusInternalServerError, errorInterface(err))
			return
		}
		s.write(c, http.StatusOK, aps)
	}
}

func (s *Server) connectToAP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.WifiHandler == nil {
			s.write(c, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		ap := models.WifiNetwork{}
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
