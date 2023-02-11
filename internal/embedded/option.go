/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/internal/embedded/models"
)

type Option func(*Handler) error

func WithHeaters(heaters map[string]Heater) Option {
	return func(h *Handler) error {
		h.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds []DSSensor) Option {
	return func(h *Handler) error {
		h.DS.sensors = make(map[string]*dsSensor)
		for _, ds := range ds {
			bus, id := ds.Name()
			cfg := ds.GetConfig()
			h.DS.sensors[id] = &dsSensor{
				DSSensor: ds,
				cfg: DSSensorConfig{
					Bus:          bus,
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
		}
		return nil
	}
}

func WithPT(pt []PTSensor) Option {
	return func(h *Handler) error {
		h.PT.sensors = make(map[string]*ptSensor)
		for _, p := range pt {
			id := p.ID()
			cfg := p.GetConfig()
			h.PT.sensors[id] = &ptSensor{
				PTSensor: p,
				PTSensorConfig: PTSensorConfig{
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
		}
		return nil
	}
}

func WithGPIOs(gpios []models.GPIO) Option {
	return func(h *Handler) error {
		h.GPIO.gpios = make(map[string]models.GPIO)
		for _, gpio := range gpios {
			h.GPIO.gpios[gpio.ID()] = gpio
		}
		return nil
	}
}

func WithLogger(l logger.Logger) Option {
	return func(*Handler) error {
		log = l
		return nil
	}
}
