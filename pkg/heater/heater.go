/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater

import (
	"errors"
	"strconv"
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

type Error struct {
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *Error) Error() string {
	if e.Err == "" {
		return "<nil>"
	}
	s := e.Op
	s += ": " + e.Err
	return s
}

var (
	ErrPowerOutOfRange = errors.New("power out of range")
	ErrNoHeating       = errors.New("lack of heating interface")
	ErrNoTicker        = errors.New("lack of ticker interface")
)

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
		if err := opt(heater); err != nil {
			return nil, err
		}
	}
	if heater.heating == nil {
		return nil, &Error{Op: "New", Err: ErrNoHeating.Error()}
	}

	if heater.ticker == nil {
		return nil, &Error{Op: "New", Err: ErrNoTicker.Error()}
	}

	if err := heater.heating.Open(); err != nil {
		return nil, &Error{Op: "New.Open", Err: err.Error()}
	}

	return heater, nil
}

func (h *Heater) Enabled() bool {
	return h.enabled.Load()
}
func (h *Heater) Power() uint {
	return h.power
}

func (h *Heater) Enable(ena bool) {
	enabled := h.Enabled()
	if ena != enabled {
		if ena {
			h.enable()
		} else {
			h.disable()
		}
	}
}

func (h *Heater) SetPower(power uint) error {
	if power > 100 {
		return &Error{Op: "SetPower " + strconv.FormatInt(int64(power), 10), Err: ErrPowerOutOfRange.Error()}
	}
	h.power = power
	return nil
}

func (h *Heater) enable() {
	h.enabled.Store(true)
	h.exit = make(chan struct{})
	h.fin = make(chan struct{})
	h.err = make(chan error, 100)

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
					err = &Error{Op: "Set", Err: err.Error()}
					select {
					case h.err <- err:
					default:
					}
				}
			}

		}
		_ = h.heating.Set(false)
		close(h.fin)
		close(h.err)
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
