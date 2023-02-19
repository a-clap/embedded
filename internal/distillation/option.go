/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type Option func(*Handler) error

func WithHeaters(heaters Heaters) Option {
	return func(h *Handler) (err error) {
		h.HeatersHandler, err = NewHandlerHeaters(heaters)
		return err
	}
}

func WithDS(ds DS) Option {
	return func(h *Handler) (err error) {
		h.DSHandler, err = NewDSHandler(ds)
		return err
	}
}

func WithPT(pt PT) Option {
	return func(h *Handler) (err error) {
		h.PTHandler, err = NewPTHandler(pt)
		return err
	}
}
