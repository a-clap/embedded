package embedded

import (
	"github.com/a-clap/iot/internal/embedded/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	Heaters *HeaterHandler
	DS      *DSHandler
	PT      *PTHandler
	GPIO    *GPIOHandler
}

var log = logger.Log

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
	{
		heaterOpts, err := parseHeaters(c.Heaters)
		if err != nil {
			log.Error("parsing ConfigHeaters resulted with errors: ", err)
		}

		if heaterOpts != nil {
			opts = append(opts, heaterOpts)
		}
	}
	{
		dsOpts, err := parseDS18B20(c.DS18B20)
		if err != nil {
			log.Error("parsing ConfigDS18B20 resulted with errors: ", err)
		}
		if dsOpts != nil {
			opts = append(opts, dsOpts)
		}
	}
	{
		ptOpts, err := parsePT100(c.PT100)
		if err != nil {
			log.Error("parsing ConfigPT100 resulted with errors: ", err)
		}
		if ptOpts != nil {
			opts = append(opts, ptOpts)
		}
	}
	{
		gpioOpts, err := parseGPIO(c.GPIO)
		if err != nil {
			log.Error("parsing ConfigGPIO resulted with errors: ", err)
		}
		if gpioOpts != nil {
			opts = append(opts, gpioOpts)
		}
	}
	return New(opts...)
}
