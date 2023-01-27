package heater

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/gpio"
)

type gpioHeating struct {
	*gpio.Out
}

func newGpioHeating(pin gpio.Pin) (*gpioHeating, error) {
	out, err := gpio.Output(pin, false)
	if err != nil {
		return nil, err
	}
	return &gpioHeating{Out: out}, nil
}

func (g *gpioHeating) Open() error {
	if g.Out == nil {
		return errors.New("gpio not usable")
	}
	return nil

}

func (g *gpioHeating) Set(b bool) error {
	return g.Out.Set(b)
}
