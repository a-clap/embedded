/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/embedded/pkg/ds18b20"
)

type DSError struct {
	ID  string `json:"ID"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (d *DSError) Error() string {
	if d.Err == "" {
		return "<nil>"
	}
	s := d.Op
	if d.ID != "" {
		s += ":" + d.ID
	}
	s += ": " + d.Err
	return s
}

type DSSensor interface {
	ID() (id string)
	Poll()
	Temperature() (actual, average float64, err error)
	GetReadings() []ds18b20.Readings
	Average() float64
	Configure(config ds18b20.SensorConfig) error
	GetConfig() ds18b20.SensorConfig
	Close()
}

type DSSensorConfig struct {
	Enabled bool `json:"enabled"`
	ds18b20.SensorConfig
}

type DSTemperature struct {
	Readings []ds18b20.Readings `json:"readings"`
}

type dsSensor struct {
	DSSensor
	cfg DSSensorConfig
}

type DSHandler struct {
	sensors map[string]*dsSensor
}

func (d *DSHandler) GetTemperatures() []DSTemperature {
	sensors := make([]DSTemperature, 0, len(d.sensors))
	
	for _, s := range d.sensors {
		tmp := DSTemperature{
			Readings: s.GetReadings(),
		}
		sensors = append(sensors, tmp)
	}
	
	return sensors
}
func (d *DSHandler) Temperature(cfg ds18b20.SensorConfig) (float64, float64, error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return 0, 0, &DSError{ID: cfg.ID, Op: "Temperature", Err: err.Error()}
	}
	actual, average, err := ds.Temperature()
	if err != nil {
		err = &DSError{ID: cfg.ID, Op: "Temperature", Err: err.Error()}
	}
	return actual, average, err
}

func (d *DSHandler) SetConfig(cfg DSSensorConfig) (newConfig DSSensorConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		err = &DSError{ID: cfg.ID, Op: "SetConfig.sensoryBy", Err: err.Error()}
		return
	}
	
	if err = ds.Configure(cfg.SensorConfig); err != nil {
		err = &DSError{ID: cfg.ID, Op: "SetConfig.Configure", Err: err.Error()}
		return
	}
	
	if cfg.Enabled != ds.cfg.Enabled {
		if cfg.Enabled {
			ds.Poll()
		} else {
			ds.Close()
		}
	}
	ds.cfg.Enabled = cfg.Enabled
	
	return d.GetConfig(cfg.ID)
}

func (d *DSHandler) GetConfig(id string) (DSSensorConfig, error) {
	s, err := d.sensorBy(id)
	if err != nil {
		return DSSensorConfig{}, &DSError{ID: id, Op: "GetConfig", Err: err.Error()}
	}
	s.cfg.SensorConfig = s.GetConfig()
	return s.cfg, nil
}

func (d *DSHandler) sensorBy(id string) (*dsSensor, error) {
	if s, ok := d.sensors[id]; ok {
		return s, nil
	}
	return nil, ErrNoSuchID
}

func (d *DSHandler) GetSensors() []DSSensorConfig {
	onewireSensors := make([]DSSensorConfig, 0, len(d.sensors))
	for _, v := range d.sensors {
		onewireSensors = append(onewireSensors, v.cfg)
	}
	return onewireSensors
}

func (d *DSHandler) Open() {
}

func (d *DSHandler) Close() {
	for _, sensor := range d.sensors {
		sensor.Close()
	}
}
