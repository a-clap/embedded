package rest_test

import (
	"github.com/gin-gonic/gin"
	"io"
)

func init() {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
}
