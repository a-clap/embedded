/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/a-clap/iot/internal/embedded/heater"
	"github.com/a-clap/iot/internal/embedded/max31865"
)

type Config struct {
	Heaters []ConfigHeater  `json:"heaters"`
	DS18B20 []ConfigDS18B20 `json:"ds18b20"`
	PT100   []ConfigPT100   `json:"pt_100"`
	GPIO    []ConfigGPIO    `json:"gpio"`
}

type ConfigHeater struct {
	string   `json:"hardware_id"`
	gpio.Pin `json:"gpio_pin"`
}

type ConfigDS18B20 struct {
	Path           string             `json:"path"`
	BusName        string             `json:"bus_name"`
	PollTimeMillis uint               `json:"poll_time_millis"`
	Resolution     ds18b20.Resolution `json:"resolution"`
	Samples        uint               `json:"samples"`
}

type ConfigPT100 struct {
	Path     string          `json:"path"`
	ID       string          `json:"id"`
	RNominal float64         `json:"r_nominal"`
	RRef     float64         `json:"r_ref"`
	Wiring   max31865.Wiring `json:"wiring"`
	ReadyPin gpio.Pin        `json:"ready_pin"`
}

type ConfigGPIO struct {
	Pin         gpio.Pin         `json:"pin"`
	ActiveLevel gpio.ActiveLevel `json:"active_level"`
	Direction   gpio.Direction   `json:"direction"`
	Value       bool             `json:"value"`
}

func parseHeaters(config []ConfigHeater) (Option, []error) {
	heaters := make(map[string]Heater, len(config))
	var errs []error
	for _, maybeHeater := range config {
		h, err := heater.New(
			heater.WithGpioHeating(maybeHeater.Pin),
			heater.WitTimeTicker(),
		)
		if err != nil {
			log.Error(err)
			errs = append(errs, err)
			continue
		}
		heaters[maybeHeater.string] = h
	}
	return WithHeaters(heaters), errs
}

func parseDS18B20(config []ConfigDS18B20) (Option, []error) {
	onewire := make([]DSSensor, len(config))
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
			onewire = append(onewire, s)
		}

	}
	return WithDS18B20(onewire), errs
}

func parsePT100(config []ConfigPT100) (Option, []error) {
	pts := make([]PTSensor, len(config))

	var errs []error
	for _, ptConfig := range config {
		s, err := max31865.NewSensor(
			max31865.WithSpidev(ptConfig.Path),
			max31865.WithID(ptConfig.ID),
			max31865.WithRNominal(ptConfig.RNominal),
			max31865.WithRefRes(ptConfig.RRef),
			max31865.WithWiring(ptConfig.Wiring),
			max31865.WithReadyPin(ptConfig.ReadyPin),
		)

		if err != nil {
			log.Errorf("error on create dsSensor: %v", err)
			errs = append(errs, err)
			continue
		}

		pts = append(pts, s)
	}

	return WithPT(pts), errs
}

func parseGPIO(config []ConfigGPIO) (Option, []error) {
	gpios := make([]GPIO, len(config))
	var errs []error
	for _, gpioConfig := range config {
		var maybeGpio *gpioHandler
		if gpioConfig.Direction == gpio.DirInput {
			gp, err := gpio.Input(gpioConfig.Pin)
			if err != nil {
				log.Errorf("error on create input: %v\n", err)
				errs = append(errs, err)
				continue
			}
			maybeGpio = &gpioHandler{GPIO: gp}
		} else {
			initValue := gpioConfig.ActiveLevel == gpio.Low
			gp, err := gpio.Output(gpioConfig.Pin, initValue)
			if err != nil {
				log.Errorf("error on create output: %v\n", err)
				errs = append(errs, err)
				continue
			}
			maybeGpio = &gpioHandler{GPIO: gp}
		}
		cfg := gpio.Config{
			ID:          "",
			Direction:   gpioConfig.Direction,
			ActiveLevel: gpioConfig.ActiveLevel,
			Value:       gpioConfig.Value,
		}

		if err := maybeGpio.Configure(cfg); err != nil {
			log.Errorf("failed to Configure on gpio: %v, err: %v", maybeGpio.ID(), err)
			errs = append(errs, err)
			continue
		}

		gpios = append(gpios, maybeGpio)
	}
	return WithGPIOs(gpios), errs
}
