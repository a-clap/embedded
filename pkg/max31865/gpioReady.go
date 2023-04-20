/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"github.com/a-clap/embedded/pkg/embedded/gpio"
	"github.com/warthog618/gpiod"
)

type gpioReady struct {
	in *gpio.In
	cb func()
}

func newGpioReady(in gpio.Pin, id string) (*gpioReady, error) {
	var err error
	m := &gpioReady{}
	m.in, err = gpio.Input(in, id, gpiod.WithPullUp, gpiod.WithFallingEdge, gpiod.WithEventHandler(m.eventHandler))
	return m, err
}

func (m *gpioReady) Open(callback func()) error {
	m.cb = callback
	return nil
}

func (m *gpioReady) Close() {
	_ = m.in.Close()
}

func (m *gpioReady) eventHandler(event gpiod.LineEvent) {
	if m.cb != nil {
		m.cb()
	}
}
