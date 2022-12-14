package max31865

import (
	"github.com/a-clap/iot/pkg/gpio"
)

type Option func(*Max) error

func WithReadWriteCloser(readWriteCloser ReadWriteCloser) Option {
	return func(max *Max) error {
		max.ReadWriteCloser = readWriteCloser
		return nil
	}
}

func WithID(id string) Option {
	return func(max *Max) error {
		max.cfg.id = id
		return nil
	}
}
func WithWiring(wiring Wiring) Option {
	return func(max *Max) error {
		max.cfg.wiring = wiring
		return nil
	}
}

func WithRefRes(res float32) Option {
	return func(max *Max) error {
		max.cfg.refRes = res
		return nil
	}
}

func WithRNominal(nominal float32) Option {
	return func(max *Max) error {
		max.cfg.rNominal = nominal
		return nil
	}
}

func WithSpidev(devfile string) Option {
	return func(max *Max) error {
		readWriteCloser, err := newMaxSpidev(devfile)
		if err == nil {
			return WithReadWriteCloser(readWriteCloser)(max)
		}
		return err
	}
}

func WithReadyPin(pin gpio.Pin) Option {
	return func(max *Max) error {
		r, err := newGpioReady(pin)
		if err == nil {
			return WithReady(r)(max)
		}
		return err
	}
}

func WithReady(r Ready) Option {
	return func(max *Max) error {
		max.cfg.ready = r
		return nil
	}
}
