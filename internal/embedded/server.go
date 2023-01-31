package embedded

import (
	"errors"
	"github.com/gin-gonic/gin"
)

const (
	RoutesGetHeaters             = "/api/heaters"
	RoutesConfigHeater           = "/api/heater/:hardware_id"
	RoutesGetOnewireSensors      = "/api/onewire"
	RoutesGetOnewireTemperatures = "/api/onewire/temperatures"
	RoutesConfigOnewireSensor    = "/api/onewire/:hardware_id"
	RoutesGetPT100Sensors        = "/api/pt100"
	RoutesGetPT100Temperatures   = "/api/pt100/temperatures"
	RoutesConfigPT100Sensor      = "/api/pt100/:hardware_id"
	RoutesGetGPIOs               = "/api/gpio"
	RoutesConfigGPIO             = "/api/gpio/:hardware_id"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

func (h *Handler) routes() {
	h.GET(RoutesGetHeaters, h.getHeaters())
	h.PUT(RoutesConfigHeater, h.configHeater())

	h.GET(RoutesGetOnewireSensors, h.getOnewireSensors())
	h.GET(RoutesGetOnewireTemperatures, h.getOnewireTemperatures())
	h.PUT(RoutesConfigOnewireSensor, h.configOnewireSensor())

	h.GET(RoutesGetPT100Sensors, h.getPTSensors())
	h.GET(RoutesGetPT100Temperatures, h.getPTTemperatures())
	h.PUT(RoutesConfigPT100Sensor, h.configPTSensor())

	h.GET(RoutesGetGPIOs, h.getGPIOS())
	h.PUT(RoutesConfigGPIO, h.configGPIO())
}

func (*Handler) respond(ctx *gin.Context, code int, obj any) {
	ctx.JSON(code, obj)
}
