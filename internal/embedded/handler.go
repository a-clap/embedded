package embedded

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	*gin.Engine
	Heaters *HeaterHandler
	DS      *DSHandler
}

func New(options ...Option) (*Handler, error) {
	h := &Handler{
		Engine:  gin.Default(),
		Heaters: new(HeaterHandler),
		DS:      new(DSHandler),
	}

	for _, opt := range options {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	h.Heaters.init()
	h.DS.Open()

	h.routes()
	return h, nil
}

func (h *Handler) Close() {

	h.DS.Close()
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
