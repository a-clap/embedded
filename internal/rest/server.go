package rest

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Server struct {
	fmt Format
	GetSensors
	*gin.Engine
}

type Format int

const (
	JSON Format = iota
	JSONP
	XML
	JSONIndent
)

func New(args ...any) *Server {
	s := &Server{
		fmt:    JSONP,
		Engine: gin.Default(),
	}

	s.parse(args...)
	s.routes()
	return s
}

func (s *Server) parse(args ...any) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case GetSensors:
			s.GetSensors = arg
		case Format:
			s.fmt = arg
		default:
			log.Printf("Unknown argument passed: {\"%T\": \"%v\"}\n", arg, arg)
		}
	}

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
