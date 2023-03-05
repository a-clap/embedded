/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

type Option func(p *Process)

func WithSensors(sensors []Sensor) Option {
	return func(p *Process) {
		for _, sensor := range sensors {
			p.sensors[sensor.ID()] = sensor
		}
	}
}

func WithHeaters(heaters []Heater) Option {
	return func(p *Process) {
		for _, heater := range heaters {
			p.heaters[heater.ID()] = heater
		}
	}
}
func WithOutputs(outputs []Output) Option {
	return func(p *Process) {
		for _, output := range outputs {
			p.outputs[output.ID()] = output
		}
	}
}

func WithClock(clock Clock) Option {
	return func(p *Process) {
		p.clock = clock
	}
}
