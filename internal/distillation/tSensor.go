/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import "github.com/a-clap/iot/internal/embedded/models"

// TSensorConfig is structure to configure each temperature sensor, no matter what type
type TSensorConfig struct {
	ID         string  `json:"id"`
	Enabled    bool    `json:"enabled"`
	Correction float32 `json:"correction"`
}

type TSensorStatus struct {
	ID          string
	Temperature float32
}

// tSensor represents temperature sensor (e.x. DS18B20, PT100)
type tSensor interface {
	Temperature() (float32, error)
	SetConfig(newCfg TSensorConfig) error
	GetConfig() (TSensorConfig, error)
}

var (
	_ tSensor = (*ptSensor)(nil)
	_ tSensor = (*dsSensor)(nil)
)

type ptSensor struct {
	sensor     models.PTSensor
	correction float32
}

type dsSensor struct {
}

func newPtSensor(sensor models.PTSensor) *ptSensor {
	return &ptSensor{sensor: sensor}
}

func (p *ptSensor) Temperature() (float32, error) {
	data := p.sensor.Temperature()
	return data.Temperature + p.correction, nil
}

func (p *ptSensor) SetConfig(newCfg TSensorConfig) error {
	//TODO implement me
	panic("implement me")
}

func (p *ptSensor) GetConfig() (TSensorConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (d *dsSensor) Temperature() (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (d *dsSensor) SetConfig(newCfg TSensorConfig) error {
	//TODO implement me
	panic("implement me")
}

func (d *dsSensor) GetConfig() (TSensorConfig, error) {
	//TODO implement me
	panic("implement me")
}
