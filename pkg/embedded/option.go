/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/logging"
)

type Option func(e *Embedded) error

func WithHeaters(heaters map[string]Heater) Option {
	return func(e *Embedded) error {
		logger.Debug("WithHeaters", logging.Int("len", len(heaters)))
		e.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds []DSSensor) Option {
	return func(e *Embedded) error {
		logger.Debug("WithDS18B20", logging.Int("len", len(ds)))
		e.DS.sensors = make(map[string]*dsSensor)
		for _, ds := range ds {
			id := ds.ID()
			cfg := ds.GetConfig()
			e.DS.sensors[id] = &dsSensor{
				DSSensor: ds,
				cfg: DSSensorConfig{
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
			logger.Debug("New DSSensor", logging.String("ID", id))
		}
		return nil
	}
}

func WithPT(pt []PTSensor) Option {
	return func(e *Embedded) error {
		logger.Debug("WithPT", logging.Int("len", len(pt)))
		e.PT.sensors = make(map[string]*ptSensor)
		for _, p := range pt {
			id := p.ID()
			cfg := p.GetConfig()
			e.PT.sensors[id] = &ptSensor{
				PTSensor: p,
				PTSensorConfig: PTSensorConfig{
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
			logger.Debug("New PTSensor", logging.String("ID", id))
		}
		return nil
	}
}

func WithGPIOs(gpios []GPIO) Option {
	return func(e *Embedded) error {
		logger.Debug("WithGPIOs", logging.Int("len", len(gpios)))
		e.GPIO.io = make(map[string]*gpioHandler)
		for _, gpio := range gpios {
			logger.Debug("New GPIO", logging.String("ID", gpio.ID()))
			e.GPIO.io[gpio.ID()] = &gpioHandler{
				GPIO: gpio}
		}
		return nil
	}
}
