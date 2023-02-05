package embedded

import (
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/internal/embedded/models"
)

type Option func(*Handler) error

type HardwareID string

func WithHeaters(heaters map[HardwareID]Heater) Option {
	return func(h *Handler) error {
		h.Heaters.heaters = heaters
		return nil
	}
}

func WithDS18B20(ds map[models.OnewireBusName][]models.DSSensor) Option {
	return func(h *Handler) error {
		h.DS.handlers = ds
		return nil
	}
}

func WithPT(pt []models.PTSensor) Option {
	return func(h *Handler) error {
		h.PT.handlers = pt
		return nil
	}
}

func WithLogger(l Logger) Option {
	return func(*Handler) error {
		log = l
		return nil
	}
}
