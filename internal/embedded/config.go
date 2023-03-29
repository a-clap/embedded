/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/a-clap/iot/internal/embedded/heater"
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/a-clap/iot/pkg/ds18b20"
)

type Config struct {
	Heaters []ConfigHeater  `mapstructure:"heaters"`
	DS18B20 []ConfigDS18B20 `mapstructure:"ds18b20"`
	PT100   []ConfigPT100   `mapstructure:"pt_100"`
	GPIO    []ConfigGPIO    `mapstructure:"gpio"`
}

type ConfigHeater struct {
	ID  string   `mapstructure:"hardware_id"`
	Pin gpio.Pin `mapstructure:"gpio_pin"`
}

type ConfigDS18B20 struct {
	Path           string             `mapstructure:"path"`
	BusName        string             `mapstructure:"bus_name"`
	PollTimeMillis uint               `mapstructure:"poll_time_millis"`
	Resolution     ds18b20.Resolution `mapstructure:"resolution"`
	Samples        uint               `mapstructure:"samples"`
}

type ConfigPT100 struct {
	Path     string          `mapstructure:"path"`
	ID       string          `mapstructure:"id"`
	RNominal float64         `mapstructure:"r_nominal"`
	RRef     float64         `mapstructure:"r_ref"`
	Wiring   max31865.Wiring `mapstructure:"wiring"`
	ReadyPin gpio.Pin        `mapstructure:"ready_pin"`
}

type ConfigGPIO struct {
	ID          string           `mapstructure:"id"`
	Pin         gpio.Pin         `mapstructure:"pin"`
	ActiveLevel gpio.ActiveLevel `mapstructure:"active_level"`
	Direction   gpio.Direction   `mapstructure:"direction"`
	Value       bool             `mapstructure:"value"`
}

func parseHeaters(config []ConfigHeater) (Option, []error) {
	log.Debugf("parsing ConfigHeater: %#v", config)

	heaters := make(map[string]Heater, len(config))
	var errs []error
	for _, maybeHeater := range config {
		h, err := heater.New(
			heater.WithGpioHeating(maybeHeater.Pin, maybeHeater.ID),
			heater.WitTimeTicker(),
		)
		if err != nil {
			log.Errorf("failed to create Heater with config %#v, err: %v ", maybeHeater, err)
			errs = append(errs, err)
			continue
		}
		heaters[maybeHeater.ID] = h
	}
	return WithHeaters(heaters), errs
}

func parseDS18B20(config []ConfigDS18B20) (Option, []error) {
	log.Debugf("parsing ConfigDS1B20: %#v", config)

	sensors := make([]DSSensor, 0, len(config))
	var errs []error
	for _, busConfig := range config {
		bus, err := ds18b20.NewBus(ds18b20.WithOnewireOnPath(busConfig.Path))
		if err != nil {
			log.Errorf("error on creating bus, path: \"%v\", error:\"%v\"\n", busConfig.Path, err)
			errs = append(errs, err)
			continue
		}

		sensors, err := bus.Discover()
		if err != nil {
			log.Error("error on discovering, err:", err)
			errs = append(errs, err)
			continue
		}
		for _, s := range sensors {
			sensors = append(sensors, s)
		}
	}
	return WithDS18B20(sensors), errs
}

func parsePT100(config []ConfigPT100) (Option, []error) {
	log.Debugf("parsing ConfigPT100: %#v", config)

	pts := make([]PTSensor, 0, len(config))
	var errs []error
	for _, cfg := range config {
		s, err := max31865.NewSensor(
			max31865.WithSpidev(cfg.Path),
			max31865.WithID(cfg.ID),
			max31865.WithRNominal(cfg.RNominal),
			max31865.WithRefRes(cfg.RRef),
			max31865.WithWiring(cfg.Wiring),
			max31865.WithReadyPin(cfg.ReadyPin, cfg.ID, nil), // TODO: inject error callback here
		)

		if err != nil {
			log.Errorf("error to create PT100 with config %#v: %v", cfg, err)
			errs = append(errs, err)
			continue
		}

		pts = append(pts, s)
	}

	return WithPT(pts), errs
}

func parseGPIO(config []ConfigGPIO) (Option, []error) {
	log.Debugf("parsing ConfigGPIO: %#v", config)

	ios := make([]GPIO, 0, len(config))
	var errs []error
	for _, cfg := range config {
		var maybeGpio *gpioHandler
		if cfg.Direction == gpio.DirInput {
			gp, err := gpio.Input(cfg.Pin, cfg.ID)
			if err != nil {
				log.Errorf("error on create input with config %#v: %v", cfg, err)
				errs = append(errs, err)
				continue
			}
			maybeGpio = &gpioHandler{GPIO: gp}
		} else {
			initValue := cfg.ActiveLevel == gpio.Low
			gp, err := gpio.Output(cfg.Pin, cfg.ID, initValue)
			if err != nil {
				log.Errorf("error on create output with config %#v: %v", cfg, err)
				errs = append(errs, err)
				continue
			}
			maybeGpio = &gpioHandler{GPIO: gp}
		}

		cfg := gpio.Config{
			ID:          "",
			Direction:   cfg.Direction,
			ActiveLevel: cfg.ActiveLevel,
			Value:       cfg.Value,
		}

		if err := maybeGpio.Configure(cfg); err != nil {
			log.Errorf("failed to Configure on gpio: %v, err: %v", maybeGpio.ID(), err)
			errs = append(errs, err)
			continue
		}

		ios = append(ios, maybeGpio)
	}
	return WithGPIOs(ios), errs
}
