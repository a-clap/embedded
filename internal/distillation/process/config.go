/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"time"
)

type Config struct {
	PhaseNumber int           `json:"phase_number"`
	Phases      []PhaseConfig `json:"phases"`
}

type PhaseConfig struct {
	Next    MoveToNextConfig    `json:"next"`
	Heaters []HeaterPhaseConfig `json:"heaters"`
	GPIO    []GPIOPhaseConfig   `json:"gpio"`
}

type MoveToNextConfig struct {
	Type                   MoveToNextType `json:"type"`
	SensorID               string         `json:"sensor_id"`
	SensorThreshold        float64        `json:"sensor_threshold"`
	TemperatureHoldSeconds int            `json:"temperature_hold_seconds"`
	SecondsToMove          int            `json:"seconds_to_move"`
}

type HeaterPhaseConfig struct {
	ID    string `json:"ID"`
	Power int    `json:"power"`
}

type GPIOPhaseConfig struct {
	ID         string  `json:"id"`
	SensorID   string  `json:"sensor_id"`
	TLow       float64 `json:"t_low"`
	THigh      float64 `json:"t_high"`
	Hysteresis float64 `json:"hysteresis"`
	Inverted   bool    `json:"inverted"`
}

type HeaterPhaseStatus struct {
	HeaterPhaseConfig
}

type GPIOPhaseStatus struct {
	ID    string `json:"id"`
	State bool   `json:"state"`
}

type MoveToNextStatusTime struct {
	TimeLeft int `json:"time_left"`
}

type MoveToNextStatusTemperature struct {
	SensorID            string  `json:"sensor_id"`
	SensorThreshold     float64 `json:"sensor_threshold"`
	TemperatureHoldLeft int     `json:"temperature_hold_left"`
}

type MoveToNextStatus struct {
	Type        MoveToNextType              `json:"type"`
	Time        MoveToNextStatusTime        `json:"time,omitempty"`
	Temperature MoveToNextStatusTemperature `json:"temperature,omitempty"`
}

type Status struct {
	Done        bool                `json:"done"`
	PhaseNumber int                 `json:"phase_number"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     time.Time           `json:"end_time"`
	Next        MoveToNextStatus    `json:"next"`
	Heaters     []HeaterPhaseStatus `json:"heaters"`
	GPIO        []GPIOPhaseStatus   `json:"gpio"`
}
