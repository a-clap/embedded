package heater

import (
	"github.com/a-clap/iot/internal/embedded/gpio"
)

type Option func(*Heater) error

func WithHeating(h Heating) Option {
	return func(heater *Heater) error {
		heater.heating = h
		return nil
	}
}

func WithTicker(t Ticker) Option {
	return func(heater *Heater) error {
		heater.ticker = t
		return nil
	}
}

func WithGpioHeating(pin gpio.Pin) Option {
	return func(heater *Heater) error {
		heating, err := newGpioHeating(pin)
		if err != nil {
			return err
		}
		heater.heating = heating
		return nil
	}
}

func WitTimeTicker() Option {
	return func(heater *Heater) error {
		heater.ticker = newTimeTicker()
		return nil
	}
}
