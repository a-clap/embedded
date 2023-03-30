/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Heater struct {
	heating Heating
	ticker  Ticker

	enabled atomic.Bool

	power        uint
	currentPower uint
	exit         chan struct{}
	fin          chan struct{}
	err          chan error
}

type Heating interface {
	Open() error
	Set(bool) error
}

type Ticker interface {
	Start(d time.Duration)
	Stop()
	Tick() <-chan time.Time
}

var (
	ErrPowerOutOfRange = errors.New("power out of range")
	ErrNoHeating       = errors.New("lack of heating interface")
	ErrNoTicker        = errors.New("lack of ticker interface")
)

// New creates new Heater with provided options
func New(options ...Option) (*Heater, error) {
	heater := &Heater{
		heating:      nil,
		ticker:       nil,
		enabled:      atomic.Bool{},
		power:        0,
		currentPower: 0,
		exit:         nil,
		fin:          nil,
		err:          nil,
	}
	for _, opt := range options {
		opt(heater)
	}

	if heater.heating == nil {
		return nil, fmt.Errorf("New: %w", ErrNoHeating)
	}

	if heater.ticker == nil {
		return nil, fmt.Errorf("New: %w", ErrNoTicker)
	}

	if err := heater.heating.Open(); err != nil {
		return nil, fmt.Errorf("New.Open: %w", err)
	}

	return heater, nil
}

// Enabled returns current state of Heater
func (h *Heater) Enabled() bool {
	return h.enabled.Load()
}

// Power returns current power of heater
func (h *Heater) Power() uint {
	return h.power
}

// Enable enables heater if it isn't enabled
func (h *Heater) Enable(err chan error) {
	if !h.Enabled() {
		h.err = err
		h.enable()
	}
}

// Disable disables heater, if it is enabled
func (h *Heater) Disable() {
	if h.Enabled() {
		h.disable()
	}
}

// SetPower set current power of heater
func (h *Heater) SetPower(power uint) error {
	if power > 100 {
		return fmt.Errorf("SetPower {Power: %v}: %w", power, ErrPowerOutOfRange)
	}
	h.power = power
	return nil
}

func (h *Heater) enable() {
	h.enabled.Store(true)
	h.exit = make(chan struct{})
	h.fin = make(chan struct{})

	loopStarted := make(chan struct{})
	go func(h *Heater) {
		h.ticker.Start(10 * time.Millisecond)
		close(loopStarted)
		for h.enabled.Load() {
			select {
			case <-h.exit:
				h.ticker.Stop()
			case <-h.ticker.Tick():
				h.currentPower = (h.currentPower + 1) % 100
				state := h.currentPower <= h.power
				if err := h.heating.Set(state); err != nil {
					// non-blocking write
					err = fmt.Errorf("Heater.Set {Value: %v}: %w", state, err)
					select {
					case h.err <- err:
					default:
					}
				}
			}

		}
		_ = h.heating.Set(false)
		close(h.fin)

		if h.err != nil {
			close(h.err)
		}
	}(h)

	// make sure loop is running
	for range loopStarted {
	}
}

func (h *Heater) disable() {
	h.enabled.Store(false)
	h.exit <- struct{}{}

	for range h.fin {
	}
	close(h.exit)
}
