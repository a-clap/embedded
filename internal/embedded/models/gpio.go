/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package models

type ActiveLevel int
type Direction int

const (
	Low ActiveLevel = iota
	High
)
const (
	Input Direction = iota
	Output
)

type GPIOConfig struct {
	ID          string      `json:"id"`
	Direction   Direction   `json:"direction"`
	ActiveLevel ActiveLevel `json:"active_level"`
	Value       bool        `json:"value"`
}

type GPIO interface {
	ID() string
	Config() (GPIOConfig, error)
	SetConfig(cfg GPIOConfig) error
}
