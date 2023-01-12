package dest

import "C"
import (
	"errors"
	"github.com/a-clap/iot/internal/embedded"
	"log"
)

type HeatersImpl interface {
	Discover() ([]embedded.HeaterConfig, error)
	Set(hardwareID string, heater embedded.HeaterConfig) error
}

var (
	ErrNoHeatersImpl = errors.New("no HeatersImpl")
)

type ConfigHeaters interface {
	HardwareIDs() ([]string, error)
	Heater(hardwareID string) (Heater, error)
}

type Heater interface {
	Config() embedded.HeaterConfig
	Enable(enable bool) error
	Power(pwr uint) error
	Name(name string) error
}

type Heaters struct {
	impl HeatersImpl
}

type heater struct {
	cfg embedded.HeaterConfig
	*Heaters
}

var (
	_ ConfigHeaters = (*Heaters)(nil)
	_ Heater        = (*heater)(nil)
)

func (h *Heaters) init() error {
	if h.impl == nil {
		return ErrNoHeatersImpl
	}

	heaters, err := h.impl.Discover()
	if err != nil {
		return err
	}

	// Disable all heaters
	for _, heater := range heaters {
		if heater.Enabled {
			heater.Enabled = false
			if err := h.impl.Set(heater.HardwareID, heater); err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}

func (h *Heaters) HardwareIDs() ([]string, error) {
	heater, err := h.getHeaters()
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(heater))
	for i, heater := range heater {
		ids[i] = heater.HardwareID
	}
	return ids, nil
}

func (h *Heaters) Heater(hardwareID string) (Heater, error) {
	found, err := h.findHeater(hardwareID)
	if err != nil {
		return nil, err
	}

	return &heater{
		cfg:     *found,
		Heaters: h,
	}, nil
}

func (h *Heaters) findHeater(hardwareID string) (*embedded.HeaterConfig, error) {
	heaters, err := h.getHeaters()
	if err != nil {
		return nil, err
	}
	for _, heater := range heaters {
		if heater.HardwareID == hardwareID {
			return &heater, nil
		}
	}
	return nil, ErrHeaterNotFound
}

func (h *Heaters) getHeaters() ([]embedded.HeaterConfig, error) {
	return h.impl.Discover()
}

func (h *heater) Config() embedded.HeaterConfig {
	return h.cfg
}

func (h *heater) Enable(enable bool) error {
	val := h.cfg
	val.Enabled = enable
	if err := h.Heaters.impl.Set(h.cfg.HardwareID, val); err != nil {
		return err
	}
	h.cfg.Enabled = enable
	return nil
}

func (h *heater) Power(pwr uint) error {
	val := h.cfg
	val.Power = pwr
	if err := h.Heaters.impl.Set(h.cfg.HardwareID, val); err != nil {
		return err
	}
	h.cfg.Power = pwr
	return nil
}

func (h *heater) Name(name string) error {
	val := h.cfg
	if err := h.Heaters.impl.Set(h.cfg.HardwareID, val); err != nil {
		return err
	}
	return nil
}
