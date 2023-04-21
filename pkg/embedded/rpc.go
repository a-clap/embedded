package embedded

import (
	"github.com/a-clap/embedded/pkg/embedded/embeddedproto"
	"github.com/a-clap/embedded/pkg/gpio"
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
