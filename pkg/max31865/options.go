package max31865

import (
	"github.com/a-clap/iot/pkg/gpio"
)

type Option func(*Sensor) error

func WithReadWriteCloser(readWriteCloser ReadWriteCloser) Option {
	return func(s *Sensor) error {
		s.ReadWriteCloser = readWriteCloser
		return nil
	}
}

func WithID(id string) Option {
	return func(s *Sensor) error {
		s.cfg.id = id
		return nil
	}
}
func WithWiring(wiring Wiring) Option {
	return func(s *Sensor) error {
		s.cfg.wiring = wiring
		return nil
	}
}

func WithRefRes(res float32) Option {
	return func(s *Sensor) error {
		s.cfg.refRes = res
		return nil
	}
}

func WithRNominal(nominal float32) Option {
	return func(s *Sensor) error {
		s.cfg.rNominal = nominal
		return nil
	}
}

func WithSpidev(devfile string) Option {
	return func(s *Sensor) error {
		readWriteCloser, err := newMaxSpidev(devfile)
		if err == nil {
			return WithReadWriteCloser(readWriteCloser)(s)
		}
		return err
	}
}

func WithReadyPin(pin gpio.Pin) Option {
	return func(s *Sensor) error {
		r, err := newGpioReady(pin)
		if err == nil {
			return WithReady(r)(s)
		}
		return err
	}
}

func WithReady(r Ready) Option {
	return func(s *Sensor) error {
		s.cfg.ready = r
		return nil
	}
}
