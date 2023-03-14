/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type Option func(*Handler) error

func WithHeaters(heaters Heaters) Option {
	return func(h *Handler) (err error) {
		if h.HeatersHandler, err = NewHandlerHeaters(heaters); err != nil {
			h.HeatersHandler = nil
		}

		return err
	}
}
func WithGPIO(gpio GPIO) Option {
	return func(h *Handler) (err error) {
		if h.GPIOHandler, err = NewGPIOHandler(gpio); err != nil {
			h.GPIOHandler = nil
		}
		return err
	}
}

func WithDS(ds DS) Option {
	return func(h *Handler) (err error) {
		if h.DSHandler, err = NewDSHandler(ds); err != nil {
			h.DSHandler = nil
		}
		return err
	}
}

func WithPT(pt PT) Option {
	return func(h *Handler) (err error) {
		if h.PTHandler, err = NewPTHandler(pt); err != nil {
			h.PTHandler = nil
		}
		return err
	}
}
