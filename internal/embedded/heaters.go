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
	HardwareID HardwareID `json:"hardware_id"`
	Enabled    bool       `json:"enabled"`
	Power      uint       `json:"power"`
}

type HeaterHandler struct {
	heaters map[HardwareID]Heater
}

func (h *Handler) configHeater() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hwid := HardwareID(ctx.Param("hardware_id"))
		if _, err := h.Heaters.StatusBy(hwid); err != nil {
			h.respond(ctx, http.StatusNotFound, err)
			return
		}

		cfg := HeaterConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		if err := h.Heaters.Config(cfg); err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		s, _ := h.Heaters.StatusBy(hwid)
		h.respond(ctx, http.StatusOK, s)
	}
}

func (h *Handler) getHeaters() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		heaters := h.Heaters.Status()
		if len(heaters) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, heaters)
	}
}

func (h *HeaterHandler) Config(cfg HeaterConfig) error {
	heater, err := h.by(HardwareID(cfg.HardwareID))
	if err != nil {
		return err
	}

	if err := heater.SetPower(cfg.Power); err != nil {
		return err
	}
	heater.Enable(cfg.Enabled)
	return nil
}

func (h *HeaterHandler) Enable(hwid HardwareID, ena bool) error {
	heat, err := h.by(hwid)
	if err != nil {
		return err
	}

	heat.Enable(ena)
	return nil
}

func (h *HeaterHandler) Power(hwid HardwareID, pwr uint) error {
	heat, err := h.by(hwid)
	if err != nil {
		return err
	}
	return heat.SetPower(pwr)
}

func (h *HeaterHandler) StatusBy(hwid HardwareID) (HeaterConfig, error) {
	heat, err := h.by(hwid)
	if err != nil {
		return HeaterConfig{}, err
	}
	return HeaterConfig{
		HardwareID: hwid,
		Enabled:    heat.Enabled(),
		Power:      heat.Power(),
	}, nil
}

func (h *HeaterHandler) Status() []HeaterConfig {
	status := make([]HeaterConfig, len(h.heaters))
	pos := 0
	for hwid, heat := range h.heaters {
		status[pos] = HeaterConfig{
			HardwareID: hwid,
			Enabled:    heat.Enabled(),
			Power:      heat.Power(),
		}
		pos++
	}
	return status
}

func (h *HeaterHandler) by(hwid HardwareID) (Heater, error) {
	maybeHeater, ok := h.heaters[hwid]
	if !ok {
		log.Debug("requested heater doesn't exist: ", hwid)
		return nil, ErrHeaterDoesntExist
	}
	return maybeHeater, nil
}

func (h *HeaterHandler) init() {
}
