/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/embedded/pkg/ds18b20"
	"github.com/a-clap/embedded/pkg/gpio"
	"github.com/a-clap/embedded/pkg/heater"
	"github.com/a-clap/embedded/pkg/max31865"
	"github.com/a-clap/logging"
)

type Config struct {
	Heaters []ConfigHeater  `mapstructure:"heaters"`
	DS18B20 []ConfigDS18B20 `mapstructure:"ds18b20"`
	PT100   []ConfigPT100   `mapstructure:"pt_100"`
	GPIO    []ConfigGPIO    `mapstructure:"gpio"`
}

type ConfigHeater struct {
	ID          string           `mapstructure:"hardware_id"`
	Pin         gpio.Pin         `mapstructure:"gpio_pin"`
	ActiveLevel gpio.ActiveLevel `mapstructure:"active_level"`
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
	Name     string          `mapstructure:"name"`
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
	logger.Debug("parseHeaters", logging.Reflect("ConfigHeater", config))

	heaters := make(map[string]Heater, len(config))
	var errs []error
	for _, maybeHeater := range config {
		h, err := heater.New(
			heater.WithGpioHeating(maybeHeater.Pin, maybeHeater.ID, maybeHeater.ActiveLevel),
			heater.WitTimeTicker(),
		)
		if err != nil {
			logger.Error("failed to create Heater ", logging.Reflect("config", maybeHeater), logging.String("error", err.Error()))
			errs = append(errs, err)
			continue
		}
		heaters[maybeHeater.ID] = h
	}
	return WithHeaters(heaters), errs
}

func parseDS18B20(config []ConfigDS18B20) (Option, []error) {
	logger.Debug("parseDS18B20", logging.Reflect("ConfigDS18B20", config))

	sensors := make([]DSSensor, 0, len(config))
	var errs []error
	for _, busConfig := range config {
		bus, err := ds18b20.NewBus(ds18b20.WithOnewireOnPath(busConfig.Path))
		if err != nil {
			logger.Error("failed to create DSBus ", logging.String("path", busConfig.Path), logging.String("error", err.Error()))
			errs = append(errs, err)
			continue
		}

		discovered, discoverErrs := bus.Discover()
		if discoverErrs != nil {
			discErrs := make([]logging.Field, len(discoverErrs)+1)
			discErrs[0] = logging.String("path", busConfig.Path)
			for i, err := range discoverErrs {
				discErrs[i+1] = logging.String("error", err.Error())
			}
			logger.Error("error on discover", discErrs...)
			errs = append(errs, discoverErrs...)
			continue
		}
		if len(discovered) == 0 {
			logger.Debug("Not found any sensors ", logging.String("path", busConfig.Path))
		}
		for _, s := range discovered {
			logger.Debug("New DSSensor", logging.String("ID", s.ID()))
			sensors = append(sensors, s)
		}
	}
	return WithDS18B20(sensors), errs
}

func parsePT100(config []ConfigPT100) (Option, []error) {
	logger.Debug("parsePT100", logging.Reflect("ConfigPT100", config))

	pts := make([]PTSensor, 0, len(config))
	var errs []error
	for _, cfg := range config {
		s, err := max31865.NewSensor(
			max31865.WithSpidev(cfg.Path),
			max31865.WithID(cfg.Path),
			max31865.WithName(cfg.Name),
			max31865.WithNominalRes(cfg.RNominal),
			max31865.WithRefRes(cfg.RRef),
			max31865.WithWiring(cfg.Wiring),
			max31865.WithReadyPin(cfg.ReadyPin, cfg.Path),
		)

		if err != nil {
			logger.Error("failed to create PT100", logging.Reflect("config", cfg), logging.String("error", err.Error()))
			errs = append(errs, err)
			continue
		}

		pts = append(pts, s)
	}

	return WithPT(pts), errs
}

func parseGPIO(config []ConfigGPIO) (Option, []error) {
	logger.Debug("parseGPIO", logging.Reflect("ConfigGPIO", config))

	ios := make([]GPIO, 0, len(config))
	var errs []error
	for _, cfg := range config {
		var maybeGpio *gpioHandler
		if cfg.Direction == gpio.DirInput {
			gp, err := gpio.Input(cfg.Pin, cfg.ID)
			if err != nil {
				logger.Error("failed to create input", logging.Reflect("config", cfg), logging.String("error", err.Error()))
				errs = append(errs, err)
				continue
			}
			maybeGpio = &gpioHandler{GPIO: gp}
		} else {
			initValue := cfg.ActiveLevel == gpio.Low
			gp, err := gpio.Output(cfg.Pin, cfg.ID, initValue)
			if err != nil {
				logger.Error("failed to create output", logging.Reflect("config", cfg), logging.String("error", err.Error()))
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
			logger.Error("failed to Configure GPIO", logging.String("ID", maybeGpio.ID()), logging.String("error", err.Error()))
			errs = append(errs, err)
			continue
		}

		ios = append(ios, maybeGpio)
	}
	return WithGPIOs(ios), errs
}
