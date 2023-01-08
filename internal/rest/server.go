package rest

import (
	"github.com/gin-gonic/gin"
)

type Server struct {
	fmt Format
	SensorHandler
	WIFIHandler
	*gin.Engine
}

type Format int

const (
	JSON Format = iota
	JSONP
	XML
	JSONIndent
)

func New(opts ...Option) (*Server, error) {
	s := &Server{
		fmt:    JSONP,
		Engine: gin.Default(),
	}

	if err := s.parse(opts...); err != nil {
		return nil, err
	}
	s.routes()
	return s, nil
}

func (s *Server) parse(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) write(c *gin.Context, code int, obj any) {
	type formatterFunc func(*gin.Context, int, any)

	var getFmt = func() map[Format]formatterFunc {
		return map[Format]formatterFunc{
			XML:        (*gin.Context).XML,
			JSON:       (*gin.Context).JSON,
			JSONP:      (*gin.Context).JSONP,
			JSONIndent: (*gin.Context).IndentedJSON,
		}
	}
	fmt := getFmt()[s.fmt]
	fmt(c, code, obj)

}
