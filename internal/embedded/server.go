package embedded

import (
	"errors"
	"github.com/gin-gonic/gin"
)

const (
	RoutesGetHeaters   = "/api/config/heaters"
	RoutesConfigHeater = "/api/config/heater/:hardware_id"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

func (h *Handler) routes() {
	h.GET(RoutesGetHeaters, h.getHeaters())
	h.PUT(RoutesConfigHeater, h.configHeater())
}

func (*Handler) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
