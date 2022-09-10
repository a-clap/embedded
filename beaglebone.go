package beaglebone

import (
	"fmt"
	"github.com/warthog618/gpiod"
	"strconv"
)
import "gobot.io/x/gobot/drivers/gpio"

// Ensure that Beaglebone fulfills needed interfaces
var _ gpio.DigitalWriter = NewAdaptor()
var _ gpio.DigitalReader = NewAdaptor()

type Beaglebone struct {
	gpio []*gpiod.Line
}

const (
	max_pin = 120
)

// NewAdaptor creates new Beaglebone, keeping consistency with Gobot
func NewAdaptor() *Beaglebone {
	return &Beaglebone{
		gpio: make([]*gpiod.Line, max_pin),
	}
}

func getPinNumber(pin string) (int, error) {
	return strconv.Atoi(pin)
}

func (b *Beaglebone) DigitalWrite(pinName string, value byte) (err error) {
	pinNumber, err := getPinNumber(pinName)
	if err != nil {
		return fmt.Errorf("can't map pin %s", pinName)
	}

	pin := b.gpio[pinNumber]
	// Create pin, if it doesn't exist
	if pin == nil {
		if b.gpio[pinNumber], err = gpiod.RequestLine("gpiochip0", pinNumber, gpiod.AsOutput(0)); err != nil {
			return fmt.Errorf("error on requesting line %v", err)
		}
		pin = b.gpio[pinNumber]
	} else {
		// Check, if pin is input
		cfg, err := pin.Info()
		if err != nil {
			return fmt.Errorf("error on info on line %v", err)
		}
		if cfg.Config.Direction == gpiod.LineDirectionInput {
			return fmt.Errorf("pin %s configured as input, can't set value", pinName)
		}
	}

	return pin.SetValue(int(value))
}

func (b *Beaglebone) DigitalRead(pinName string) (val int, err error) {
	pinNumber, err := getPinNumber(pinName)
	if err != nil {
		return 0, fmt.Errorf("can't map pin %s", pinName)
	}
	pin := b.gpio[pinNumber]
	if pin == nil {
		if b.gpio[pinNumber], err = gpiod.RequestLine("gpiochip0", pinNumber, gpiod.AsInput); err != nil {
			return 0, fmt.Errorf("error on requesting line %v", err)
		}
		pin = b.gpio[pinNumber]
	}
	return pin.Value()
}
