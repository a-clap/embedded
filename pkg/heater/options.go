/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater

import (
	"github.com/a-clap/embedded/pkg/gpio"
)

type Option func(*Heater)

func WithHeating(h Heating) Option {
	return func(heater *Heater) {
		heater.heating = h
	}
}

func WithTicker(t Ticker) Option {
	return func(heater *Heater) {
		heater.ticker = t
	}
}

func WithGpioHeating(pin gpio.Pin, id string, level gpio.ActiveLevel) Option {
	return func(heater *Heater) {
		heater.heating = newGpioHeating(pin, id, level)
	}
}

func WitTimeTicker() Option {
	return func(heater *Heater) {
		heater.ticker = newTimeTicker()
	}
}
