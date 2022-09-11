package beaglebone

import (
	"errors"
	"fmt"
	"github.com/warthog618/gpiod"
	"strconv"
)

type gpio struct {
	gpio []*gpiod.Line
}

var (
	ErrCantParsePin = errors.New("can't parse pin")
)

// There is no need to create multiple gpio handlers
var gpioHandler *gpio = nil

// newGpio creates gpioHandler or returns existing one
func newGpio() *gpio {
	if gpioHandler == nil {
		cc := gpiod.Chips()
		Log.Debug("found chips: ", cc)
		pins := 0

		for _, c := range cc {
			chip, err := gpiod.NewChip(c)
			if err != nil {
				Log.Warnf("err on requesting chip %v", c)
			}
			pins += chip.Lines()
			if err := chip.Close(); err != nil {
				Log.Errorf("error on closing chip %v: %v", chip, err)
			}
		}

		Log.Debug("detected gpio pins: ", pins)
		gpioHandler = &gpio{
			gpio: make([]*gpiod.Line, pins),
		}
	}
	return gpioHandler
}

func (g *gpio) DigitalWrite(pinName string, value byte) (err error) {
	chip, pinNumber, err := g.parsePin(pinName)
	if err != nil {
		return ErrCantParsePin
	}

	pin := g.gpio[pinNumber]
	// Create pin, if it doesn't exist
	if pin == nil {
		if g.gpio[pinNumber], err = gpiod.RequestLine(chip, pinNumber, gpiod.AsOutput(0)); err != nil {
			return fmt.Errorf("error on requesting line %v", err)
		}
		pin = g.gpio[pinNumber]
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

func (g *gpio) DigitalRead(pinName string) (val int, err error) {
	chip, pinNumber, err := g.parsePin(pinName)
	if err != nil {
		return 0, fmt.Errorf("can't map pin %s", pinName)
	}

	pin := g.gpio[pinNumber]
	if pin == nil {
		if g.gpio[pinNumber], err = gpiod.RequestLine(chip, pinNumber, gpiod.AsInput); err != nil {
			return 0, fmt.Errorf("error on requesting line %v", err)
		}
		pin = g.gpio[pinNumber]
	}
	return pin.Value()
}

func (g *gpio) parsePin(pin string) (chip string, pinNumber int, err error) {
	pinNumber, err = strconv.Atoi(pin)
	if err != nil {
		Log.Warnf("failed strconv on %v", pin)
		return
	}
	// Each gpiochip has at most 32 pins (0 - 31)
	chip = "gpiochip" + strconv.FormatInt(int64(pinNumber/32), 10)

	pinNumber %= 32
	Log.Debugf("getPinNumber: for pin %v, chip: %v, pinNumber: %v", pin, chip, pinNumber)
	return
}
