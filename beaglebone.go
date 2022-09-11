package beaglebone

import (
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	gobotgpio "gobot.io/x/gobot/drivers/gpio"
)

var Log logger.Logger = logger.NewDefaultZap(zapcore.ErrorLevel)

// Ensure that Beaglebone fulfills needed interfaces
var _ gobotgpio.DigitalWriter = NewAdaptor()
var _ gobotgpio.DigitalReader = NewAdaptor()

type Beaglebone struct {
	*gpio
}

// NewAdaptor creates new Beaglebone, keeping consistency with Gobot
func NewAdaptor() *Beaglebone {
	return &Beaglebone{
		gpio: newGpio(),
	}
}
