/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"fmt"

	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/warthog618/gpiod"
)

type gpioReady struct {
	pin         gpio.Pin
	in          *gpio.In
	cb          func(any) error
	cbArgs      any
	errCallback func(error)
}

func newGpioReady(in gpio.Pin, id string, errCallback func(err error)) (*gpioReady, error) {
	var err error
	m := &gpioReady{pin: in}
	m.in, err = gpio.Input(m.pin, id, gpiod.WithPullUp, gpiod.WithFallingEdge, gpiod.WithEventHandler(m.eventHandler))
	return m, err
}

func (m *gpioReady) Open(callback func(any) error, args any) error {
	m.cb = callback
	m.cbArgs = args

	return nil
}

func (m *gpioReady) Close() {
	_ = m.in.Close()
}

func (m *gpioReady) eventHandler(event gpiod.LineEvent) {
	if m.cb != nil {
		if err := m.cb(m.cbArgs); err != nil && m.errCallback != nil {
			m.errCallback(fmt.Errorf("%w: happened on event %v", err, event))
		}
	}
}
