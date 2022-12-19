package embedded

import (
	"github.com/a-clap/iot/pkg/gpio"
	"github.com/a-clap/iot/pkg/heater"
)

type Config struct {
	Heaters []ConfigHeater `json:"heaters"`
}

type ConfigHeater struct {
	HardwareID `json:"hardware_id"`
	gpio.Pin   `json:"gpioPin"`
}

func parseHeaters(config []ConfigHeater) (Option, []error) {
	heaters := make(map[HardwareID]Heater, len(config))
	var errs []error
	for _, maybeHeater := range config {
		h, err := heater.New(
			heater.WithGpioHeating(maybeHeater.Pin),
			heater.WitTimeTicker(),
		)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		heaters[maybeHeater.HardwareID] = h
	}
	return WithHeaters(heaters), errs
}
