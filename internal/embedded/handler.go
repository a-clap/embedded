package embedded

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	heaters map[HardwareID]Heater
}

func New(options ...Option) (*Handler, error) {
	h := &Handler{
		Engine: gin.Default(),
	}
	for _, opt := range options {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	h.routes()
	return h, nil
}

func NewFromConfig(c Config) (*Handler, error) {
	var opts []Option

	heaterOpts, err := parseHeaters(c.Heaters)
	if err != nil {
		log.Error("parsing heaterOpts resulted with errors: ", err)
	}
	if heaterOpts != nil {
		opts = append(opts, heaterOpts)
	}

	return New(opts...)
}
