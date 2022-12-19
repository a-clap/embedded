package embedded

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Heater interface {
	Enable(ena bool)
	SetPower(pwr uint) error
	Enabled() bool
	Power() uint
}

type HeaterConfig struct {
	HardwareID string `json:"hardware_id"`
	Enabled    bool   `json:"enabled"`
	Power      uint   `json:"power"`
}

func (h *Handler) configHeater() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hwid := ctx.Param("hardware_id")
		heater, err := h.heaterBy(HardwareID(hwid))

		if err != nil {
			h.respond(ctx, http.StatusNotFound, ErrHeaterDoesntExist)
			return
		}

		cfg := HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		if err := heater.SetPower(cfg.Power); err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		heater.Enable(cfg.Enabled)
		s, err := h.HeaterStatusBy(HardwareID(hwid))
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		h.respond(ctx, http.StatusOK, s)

	}
}

func (h *Handler) getHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(h.heaters) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}

		heaters := make([]HeaterConfig, len(h.heaters))
		var i int
		for k, v := range h.heaters {
			heaters[i].HardwareID = string(k)
			heaters[i].Enabled = v.Enabled()
			heaters[i].Power = v.Power()
			i++
		}

		h.respond(ctx, http.StatusOK, heaters)

	}
}

func (h *Handler) HeaterEnable(hwid HardwareID, ena bool) error {
	heat, err := h.heaterBy(hwid)
	if err != nil {
		return err
	}

	heat.Enable(ena)
	return nil
}

func (h *Handler) HeaterPower(hwid HardwareID, pwr uint) error {
	heat, err := h.heaterBy(hwid)
	if err != nil {
		return err
	}
	return heat.SetPower(pwr)
}

func (h *Handler) HeaterStatusBy(hwid HardwareID) (HeaterConfig, error) {
	heat, err := h.heaterBy(hwid)
	if err != nil {
		return HeaterConfig{}, err
	}
	return HeaterConfig{
		HardwareID: string(hwid),
		Enabled:    heat.Enabled(),
		Power:      heat.Power(),
	}, nil
}

func (h *Handler) HeatersStatus() []HeaterConfig {
	status := make([]HeaterConfig, len(h.heaters))
	pos := 0
	for key, heat := range h.heaters {
		status[pos] = HeaterConfig{
			HardwareID: string(key),
			Enabled:    heat.Enabled(),
			Power:      heat.Power(),
		}
		pos++
	}
	return status
}

func (h *Handler) heaterBy(hwid HardwareID) (Heater, error) {
	maybeHeater, ok := h.heaters[hwid]
	if !ok {
		log.Debug("requested heater doesn't exist: ", hwid)
		return nil, ErrHeaterDoesntExist
	}
	return maybeHeater, nil
}
