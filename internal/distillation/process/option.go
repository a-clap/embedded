/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

type Option func(p *Process)

func WithSensors(sensors []Sensor) Option {
	return func(p *Process) {
		p.ConfigureSensors(sensors)
	}
}

func WithHeaters(heaters []Heater) Option {
	return func(p *Process) {
		p.ConfigureHeaters(heaters)
	}
}
func WithOutputs(outputs []Output) Option {
	return func(p *Process) {
		p.ConfigureOutputs(outputs)
	}
}

func WithClock(clock Clock) Option {
	return func(p *Process) {
		p.clock = clock
	}
}
