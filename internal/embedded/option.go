package embedded

import (
	"github.com/a-clap/iot/internal/embedded/dsSensor"
	. "github.com/a-clap/iot/internal/embedded/logger"
)

type Option func(*Handler) error

type HardwareID string

func WithHeaters(heaters map[HardwareID]Heater) Option {
	return func(h *Handler) error {
		h.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds map[OnewireBusName][]dsSensor.Handler) Option {
	return func(h *Handler) error {
		h.DS.sensors = ds
		return nil
	}
}

func WithLogger(l Logger) Option {
	return func(*Handler) error {
		Log = l
		return nil
	}
}
