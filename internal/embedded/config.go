package embedded

import (
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/a-clap/iot/pkg/gpio"
	"github.com/a-clap/iot/pkg/heater"
)

type Config struct {
	Heaters []ConfigHeater `json:"heaters"`
}

type ConfigHeater struct {
	HardwareID `json:"hardware_id"`
	gpio.Pin   `json:"gpio_pin"`
}

type ConfigDS18B20 struct {
	BusName        models.OnewireBusName `json:"bus_name"`
	PollTimeMillis uint                  `json:"poll_time_millis"`
	Resolution     models.DSResolution   `json:"resolution"`
	Samples        uint                  `json:"samples"`
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

//func parseDS18B20(Config []ConfigDS18B20) (Option, error) {
//	ds := make(map[models.OnewireBusName][]DSSensorHandler)
//	var errs []error
//
//	for _, maybeOnewire := range Config {
//
//	}
//
//}
//
