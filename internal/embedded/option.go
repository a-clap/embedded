/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

type Option func(*Handler) error

func WithHeaters(heaters map[string]Heater) Option {
	return func(h *Handler) error {
		log.Debug("with heaters: ", len(heaters))
		h.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds []DSSensor) Option {
	return func(h *Handler) error {
		log.Debug("with ds18b20")
		h.DS.sensors = make(map[string]*dsSensor)
		for _, ds := range ds {
			bus, id := ds.Name()
			log.Debug("Adding ds: ", bus, id)
			cfg := ds.GetConfig()
			h.DS.sensors[id] = &dsSensor{
				DSSensor: ds,
				cfg: DSSensorConfig{
					Bus:          bus,
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
			log.Debugf("new dsSensor on bus: %v with ID: %v", bus, id)
		}
		return nil
	}
}

func WithPT(pt []PTSensor) Option {
	return func(h *Handler) error {
		log.Debug("with pt100s")
		h.PT.sensors = make(map[string]*ptSensor)
		for _, p := range pt {
			log.Debug("Adding pt100 ", p.ID())
			id := p.ID()
			cfg := p.GetConfig()
			h.PT.sensors[id] = &ptSensor{
				PTSensor: p,
				PTSensorConfig: PTSensorConfig{
					Enabled:      false,
					SensorConfig: cfg,
				},
			}
			log.Debugf("new ptSensor with ID: %v", id)
		}
		return nil
	}
}

func WithGPIOs(gpios []GPIO) Option {
	return func(h *Handler) error {
		h.GPIO.io = make(map[string]*gpioHandler)
		for _, gpio := range gpios {
			log.Debug("adding GPIO: ", gpio.ID())
			h.GPIO.io[gpio.ID()] = &gpioHandler{
				GPIO: gpio}
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
