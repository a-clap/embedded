/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"github.com/a-clap/embedded/pkg/gpio"
)

type Option func(*Sensor) error

// WithReadWriteCloser sets interface to communicate with Max
func WithReadWriteCloser(readWriteCloser ReadWriteCloser) Option {
	return func(s *Sensor) error {
		s.ReadWriteCloser = readWriteCloser
		return nil
	}
}

// WithName sets Max name - it can be changed
func WithName(name string) Option {
	return func(s *Sensor) error {
		s.cfg.Name = name
		return nil
	}
}

// WithID sets unique ID for Sensor - cannot be changed
func WithID(id string) Option {
	return func(s *Sensor) error {
		s.cfg.ID = id
		return nil
	}
}

// WithWiring sets sensor wiring
func WithWiring(wiring Wiring) Option {
	return func(s *Sensor) error {
		s.configReg.wiring = wiring
		return nil
	}
}

// WithRefRes sets value of reference resistor on pcb
func WithRefRes(res float64) Option {
	return func(s *Sensor) error {
		s.configReg.refRes = res
		return nil
	}
}

// WithNominalRes sets value of nominal resistance (usually 100)
func WithNominalRes(nominal float64) Option {
	return func(s *Sensor) error {
		s.configReg.nominalRes = nominal
		return nil
	}
}

// WithSpidev is a standard way of communication with Max - via spidev
func WithSpidev(devfile string) Option {
	return func(s *Sensor) error {
		readWriteCloser, err := newMaxSpidev(devfile)
		if err == nil {
			return WithReadWriteCloser(readWriteCloser)(s)
		}
		return err
	}
}

// WithReadyPin returns interface for async Poll on Max - based on DRDY pin
func WithReadyPin(pin gpio.Pin, id string) Option {
	return func(s *Sensor) error {
		r, err := newGpioReady(pin, id)
		if err == nil {
			return WithReady(r)(s)
		}
		return err
	}
}

// WithReady returns user Ready interface for async Poll
func WithReady(r Ready) Option {
	return func(s *Sensor) error {
		s.ready = r
		return nil
	}
}
