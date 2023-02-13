/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"github.com/a-clap/iot/internal/embedded/gpio"
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
		s.configReg.id = id
		return nil
	}
}
func WithWiring(wiring Wiring) Option {
	return func(s *Sensor) error {
		s.configReg.wiring = wiring
		return nil
	}
}

func WithRefRes(res float32) Option {
	return func(s *Sensor) error {
		s.configReg.refRes = res
		return nil
	}
}

func WithRNominal(nominal float32) Option {
	return func(s *Sensor) error {
		s.configReg.rNominal = nominal
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
		s.ready = r
		return nil
	}
}
