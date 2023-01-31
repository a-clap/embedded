package embedded

import (
	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/a-clap/iot/internal/embedded/heater"
	"github.com/a-clap/iot/internal/embedded/models"
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
