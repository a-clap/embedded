package embedded

import (
	"time"
	
	"github.com/a-clap/embedded/pkg/ds18b20"
	"github.com/a-clap/embedded/pkg/embedded/embeddedproto"
	"github.com/a-clap/embedded/pkg/gpio"
	"github.com/a-clap/embedded/pkg/max31865"
)

func gpioConfigToRPC(config *GPIOConfig) *embeddedproto.GPIOConfig {
	return &embeddedproto.GPIOConfig{
		ID:          config.ID,
		Direction:   int32(config.Direction),
		ActiveLevel: int32(config.ActiveLevel),
		Value:       config.Value,
	}
}

func rpcToGPIOConfig(config *embeddedproto.GPIOConfig) GPIOConfig {
	return GPIOConfig{gpio.Config{
		ID:          config.ID,
		Direction:   gpio.Direction(config.Direction),
		ActiveLevel: gpio.ActiveLevel(config.ActiveLevel),
		Value:       config.Value,
	}}
}

func rpcToDSConfig(elem *embeddedproto.DSConfig) DSSensorConfig {
	return DSSensorConfig{
		Enabled: elem.Enabled,
		SensorConfig: ds18b20.SensorConfig{
			Name:         elem.Name,
			ID:           elem.ID,
			Correction:   float64(elem.Correction),
			Resolution:   ds18b20.Resolution(elem.Resolution),
			PollInterval: time.Duration(elem.PollInterval),
			Samples:      uint(elem.Samples),
		},
	}
}

func dsConfigToRPC(d *DSSensorConfig) *embeddedproto.DSConfig {
	return &embeddedproto.DSConfig{
		ID:           d.ID,
		Name:         d.Name,
		Correction:   float32(d.Correction),
		Resolution:   int32(d.Resolution),
		PollInterval: int32(d.PollInterval),
		Samples:      uint32(d.Samples),
		Enabled:      d.Enabled,
	}
}

func rpcToDSTemperature(r *embeddedproto.DSTemperatures) []DSTemperature {
	temperatures := make([]DSTemperature, len(r.Temps))
	for i, temp := range r.Temps {
		readings := make([]ds18b20.Readings, len(temp.Readings))
		for j, r := range temp.Readings {
			readings[j] = ds18b20.Readings{
				ID:          r.ID,
				Temperature: float64(r.Temperature),
				Average:     float64(r.Average),
				Stamp:       time.UnixMilli(r.StampMillis),
				Error:       r.Error,
			}
		}
		temperatures[i] = DSTemperature{Readings: readings}
	}
	return temperatures
}

func dsTemperatureToRPC(t []DSTemperature) *embeddedproto.DSTemperatures {
	temperatures := &embeddedproto.DSTemperatures{}
	temperatures.Temps = make([]*embeddedproto.DSTemperature, len(t))
	for i, temp := range t {
		readings := make([]*embeddedproto.DSReadings, len(temp.Readings))
		for j, r := range temp.Readings {
			readings[j] = &embeddedproto.DSReadings{
				ID:          r.ID,
				Temperature: float32(r.Temperature),
				Average:     float32(r.Average),
				StampMillis: r.Stamp.UnixMilli(),
				Error:       r.Error,
			}
		}
		temperatures.Temps[i] = &embeddedproto.DSTemperature{Readings: readings}
	}
	return temperatures
}

func rpcToPTConfig(elem *embeddedproto.PTConfig) PTSensorConfig {
	return PTSensorConfig{
		Enabled: elem.Enabled,
		SensorConfig: max31865.SensorConfig{
			Name:         elem.Name,
			ID:           elem.ID,
			Correction:   float64(elem.Correction),
			ASyncPoll:    elem.Async,
			PollInterval: time.Duration(elem.PollInterval),
			Samples:      uint(elem.Samples),
		},
	}
}

func ptConfigToRPC(d *PTSensorConfig) *embeddedproto.PTConfig {
	return &embeddedproto.PTConfig{
		ID:           d.ID,
		Name:         d.Name,
		Correction:   float32(d.Correction),
		Async:        d.ASyncPoll,
		PollInterval: int32(d.PollInterval),
		Samples:      uint32(d.Samples),
		Enabled:      d.Enabled,
	}
}

func rpcToPTTemperature(r *embeddedproto.PTTemperatures) []PTTemperature {
	temperatures := make([]PTTemperature, len(r.Temps))
	for i, temp := range r.Temps {
		readings := make([]max31865.Readings, len(temp.Readings))
		for j, r := range temp.Readings {
			readings[j] = max31865.Readings{
				ID:          r.ID,
				Temperature: float64(r.Temperature),
				Average:     float64(r.Average),
				Stamp:       time.UnixMilli(r.StampMillis),
				Error:       r.Error,
			}
		}
		temperatures[i] = PTTemperature{Readings: readings}
	}
	return temperatures
}

func ptTemperatureToRPC(t []PTTemperature) *embeddedproto.PTTemperatures {
	temperatures := &embeddedproto.PTTemperatures{}
	temperatures.Temps = make([]*embeddedproto.PTTemperature, len(t))
	for i, temp := range t {
		readings := make([]*embeddedproto.PTReadings, len(temp.Readings))
		for j, r := range temp.Readings {
			readings[j] = &embeddedproto.PTReadings{
				ID:          r.ID,
				Temperature: float32(r.Temperature),
				Average:     float32(r.Average),
				StampMillis: r.Stamp.UnixMilli(),
				Error:       r.Error,
			}
		}
		temperatures.Temps[i] = &embeddedproto.PTTemperature{Readings: readings}
	}
	return temperatures
}
