package embedded

import (
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	Heaters *HeaterHandler
	DS      *DSHandler
	PT      *PTHandler
	GPIO    *GPIOHandler
}

var log = Log

func New(options ...Option) (*Handler, error) {
	h := &Handler{
		Engine:  gin.Default(),
		Heaters: new(HeaterHandler),
		DS:      new(DSHandler),
		PT:      new(PTHandler),
		GPIO:    new(GPIOHandler),
	}

	for _, opt := range options {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	h.Heaters.Open()
	h.DS.Open()
	h.PT.Open()
	h.GPIO.Open()

	h.routes()
	return h, nil
}

func (h *Handler) Close() {
	h.Heaters.Close()
	h.DS.Close()
	h.PT.Close()
	h.GPIO.Close()
}

func NewFromConfig(c Config) (*Handler, error) {
	var opts []Option

	heaterOpts, err := parseHeaters(c.Heaters)
	if err != nil {
		log.Error("parsing ConfigHeaters resulted with errors: ", err)
	}
	if heaterOpts != nil {
		opts = append(opts, heaterOpts)
	}

	return New(opts...)
}
