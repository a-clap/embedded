package embedded

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	ErrNoSuchSensor = errors.New("specified sensor doesnt' exist")
)

type DSHandler struct {
	handlers map[models.OnewireBusName][]models.DSSensor
	sensors  map[string]models.DSSensor
}

func (h *Handler) configOnewireSensor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hwid := ctx.Param("hardware_id")
		if _, err := h.DS.GetConfig(hwid); err != nil {
			h.respond(ctx, http.StatusNotFound, err)
			return
		}

		cfg := models.DSConfig{}
		if err := ctx.ShouldBind(&cfg); err != nil {
			h.respond(ctx, http.StatusBadRequest, err)
			return
		}

		cfg, err := h.DS.SetConfig(cfg)
		if err != nil {
			h.respond(ctx, http.StatusInternalServerError, toError(err))
			return
		}

		h.respond(ctx, http.StatusOK, cfg)
	}
}
func (h *Handler) getOnewireTemperatures() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ds := h.DS.GetTemperatures()
		if len(ds) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, ds)
	}
}

func (h *Handler) getOnewireSensors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ds := h.DS.GetSensors()
		if len(ds) == 0 {
			h.respond(ctx, http.StatusInternalServerError, ErrNotImplemented)
			return
		}
		h.respond(ctx, http.StatusOK, ds)
	}
}

func (d *DSHandler) GetTemperatures() []models.Temperature {
	sensors := make([]models.Temperature, 0, len(d.sensors))

	for _, sensor := range d.sensors {
		sensors = append(sensors, sensor.Temperature())
	}

	return sensors
}
func (d *DSHandler) SensorTemperature(cfg models.DSConfig) (status models.Temperature, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	return ds.Temperature(), nil
}

func (d *DSHandler) SetConfig(cfg models.DSConfig) (newConfig models.DSConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = ds.SetConfig(cfg); err != nil {
		return
	}

	return ds.Config(), nil
}

func (d *DSHandler) GetConfig(id string) (models.DSConfig, error) {
	s, err := d.sensorBy(id)
	if err != nil {
		return models.DSConfig{}, err
	}
	return s.Config(), nil
}

func (d *DSHandler) sensorBy(id string) (models.DSSensor, error) {
	if s, ok := d.sensors[id]; ok {
		return s, nil
	}
	return nil, ErrNoSuchSensor
}

func (d *DSHandler) GetSensors() []models.OnewireSensors {
	onewireSensors := make([]models.OnewireSensors, len(d.handlers))

	pos := 0
	for k, v := range d.handlers {
		onewireSensors[pos].Bus = k
		onewireSensors[pos].DSConfig = make([]models.DSConfig, 0, len(v))
		for _, sensor := range v {
			cfg := sensor.Config()
			if _, ok := d.sensors[cfg.ID]; ok {
				onewireSensors[pos].DSConfig = append(onewireSensors[pos].DSConfig, cfg)
			} else {
				log.Debug("id not found before: ", cfg.ID)
			}
		}
		pos++
	}
	return onewireSensors
}

func (d *DSHandler) Open() {
	if d.handlers == nil {
		return
	}

	d.sensors = make(map[string]models.DSSensor)
	for _, sensors := range d.handlers {
		for _, sensor := range sensors {
			cfg := sensor.Config()
			d.sensors[cfg.ID] = sensor
		}
	}
}

func (d *DSHandler) Close() {
	for name, sensor := range d.sensors {
		cfg := sensor.Config()
		if cfg.Enabled {
			cfg.Enabled = false
			err := sensor.SetConfig(cfg)
			if err != nil {
				log.Debug("SetConfig failed on sensor: ", name)
			}
		}
	}
}
