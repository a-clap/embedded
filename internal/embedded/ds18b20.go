package embedded

import (
	"errors"
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/internal/embedded/models"
)

type OnewireBusName string

var (
	ErrNoSuchSensor = errors.New("specified sensor doesnt' exist")
)

type OnewireSensors struct {
	Bus      OnewireBusName    `json:"bus"`
	DSConfig []models.DSConfig `json:"ds18b20"`
}

type DSHandler struct {
	handlers map[OnewireBusName][]models.DSSensor
	sensors  map[string]models.DSSensor
}

func (d *DSHandler) ConfigSensor(cfg models.DSConfig) (newConfig models.DSConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = ds.SetConfig(cfg); err != nil {
		return
	}

	return ds.Config(), nil
}

func (d *DSHandler) SensorStatus(id string) (models.DSConfig, error) {
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

func (d *DSHandler) Status() ([]OnewireSensors, error) {
	onewireSensors := make([]OnewireSensors, len(d.handlers))

	pos := 0
	for k, v := range d.handlers {
		onewireSensors[pos].Bus = k
		onewireSensors[pos].DSConfig = make([]models.DSConfig, 0, len(v))
		for _, sensor := range v {
			cfg := sensor.Config()
			if _, ok := d.sensors[cfg.ID]; ok {
				onewireSensors[pos].DSConfig = append(onewireSensors[pos].DSConfig, cfg)
			} else {
				Log.Debug("id not found before: ", cfg.ID)
			}
		}
		pos++
	}
	return onewireSensors, nil
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
	for _, s := range d.sensors {
		_ = s.StopPoll()
	}
}
