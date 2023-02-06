/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package models

type PTConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Samples uint   `json:"samples"`
}

type PTSensor interface {
	ID() string
	Temperature() Temperature
	Config() PTConfig
	SetConfig(cfg PTConfig) (err error)
}