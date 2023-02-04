/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package models

type HeaterID string

type Heater interface {
	Enable(ena bool)
	SetPower(pwr uint) error
	Enabled() bool
	Power() uint
}

type HeaterConfig struct {
	HardwareID HeaterID `json:"hardware_id"`
	Enabled    bool     `json:"enabled"`
	Power      uint     `json:"power"`
}
