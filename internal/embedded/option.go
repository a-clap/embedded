/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/internal/embedded/models"
)

type Option func(*Handler) error

func WithHeaters(heaters map[string]models.Heater) Option {
	return func(h *Handler) error {
		h.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds []DSSensor) Option {
	return func(h *Handler) error {
		h.DS.sensors = make(map[string]*sensor)
		for _, ds := range ds {
			bus, id := ds.Name()
			cfg := ds.GetConfig()
			h.DS.sensors[id] = &sensor{
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

func WithPT(pt []models.PTSensor) Option {
	return func(h *Handler) error {
		h.PT.handlers = pt
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

func WithLogger(l Logger) Option {
	return func(*Handler) error {
		log = l
		return nil
	}
}
