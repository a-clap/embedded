/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater

import (
	"github.com/a-clap/iot/pkg/embedded/gpio"
)

type gpioHeating struct {
	*gpio.Out
	err error
}

var _ Heating = (*gpioHeating)(nil)

func newGpioHeating(pin gpio.Pin, id string) *gpioHeating {
	out, err := gpio.Output(pin, id, false)
	return &gpioHeating{Out: out, err: err}
}

func (g *gpioHeating) Open() error {
	return g.err
}

func (g *gpioHeating) Set(b bool) error {
	return g.Out.Set(b)
}
