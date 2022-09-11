package beaglebone

import (
	"fmt"
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	gobotgpio "gobot.io/x/gobot/drivers/gpio"
)

var Log logger.Logger = logger.NewDefaultZap(zapcore.ErrorLevel)

// Ensure that Beaglebone fulfills needed interfaces
var _ gobotgpio.DigitalWriter = NewAdaptor()
var _ gobotgpio.DigitalReader = NewAdaptor()

type Beaglebone struct {
	name string
	*gpio
}

// NewAdaptor creates new Beaglebone, keeping consistency with Gobot
func NewAdaptor() *Beaglebone {
	return &Beaglebone{
		gpio: newGpio(),
	}
}

// Name fulfils gobot.Adaptor interface
func (b *Beaglebone) Name() string {
	return b.name
}

// SetName fulfils gobot.Adaptor interface
func (b *Beaglebone) SetName(name string) {
	b.name = name
}

// Connect fulfils gobot.Adaptor interface
func (b *Beaglebone) Connect() error {
	// not sure, if we should do anything for now
	return nil
}

// Finalize fulfils gobot.Adaptor interface
func (b *Beaglebone) Finalize() error {
	err := b.gpio.Finalize()
	if err != nil {
		Log.Error(err)
		return fmt.Errorf("error on gpio close")
	}
	return nil
}
