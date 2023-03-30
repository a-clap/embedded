/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embeddedmock

type Heater struct {
	enabled bool
	pwr     uint
}

func NewHeater() *Heater {
	return &Heater{
		enabled: false,
		pwr:     0,
	}
}

func (h *Heater) Enable(err chan error) {
	h.enabled = true
}

func (h *Heater) Disable() {
	h.enabled = false
}

func (h *Heater) SetPower(pwr uint) error {
	h.pwr = pwr
	return nil
}

func (h *Heater) Enabled() bool {
	return h.enabled
}

func (h *Heater) Power() uint {
	return h.pwr
}
