package embedded

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/dsSensor"
	. "github.com/a-clap/iot/internal/embedded/logger"
)

type OnewireBusName string

var (
	ErrNoSuchSensor = errors.New("specified handler doesnt' exist")
)

type BusConfig struct {
	Resolution     dsSensor.Resolution `json:"resolution"`
	PollTimeMillis uint                `json:"poll_time_millis"`
	Samples        uint                `json:"samples"`
}

type OnewireSensors struct {
	Bus      OnewireBusName    `json:"bus"`
	DSConfig []dsSensor.Config `json:"ds18b20"`
}

type DSHandler struct {
	sensors map[OnewireBusName][]dsSensor.Handler
	cfg     map[string]*dsSensor.Sensor
}

func (d *DSHandler) ConfigSensor(cfg dsSensor.Config) (newConfig dsSensor.Config, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = ds.SetConfig(cfg); err != nil {
		return
	}

	return ds.Config(), nil
}

func (d *DSHandler) SensorStatus(id string) (dsSensor.Config, error) {
	s, err := d.sensorBy(id)
	if err != nil {
		return dsSensor.Config{}, err
	}
	return s.Config(), nil
}

func (d *DSHandler) sensorBy(id string) (*dsSensor.Sensor, error) {
	if s, ok := d.cfg[id]; ok {
		return s, nil
	}
	return nil, ErrNoSuchSensor
}

func (d *DSHandler) Status() ([]OnewireSensors, error) {
	onewireSensors := make([]OnewireSensors, len(d.sensors))

	pos := 0
	for k, v := range d.sensors {
		onewireSensors[pos].Bus = k
		onewireSensors[pos].DSConfig = make([]dsSensor.Config, 0, len(v))
		for _, sensor := range v {
			id := sensor.ID()
			if cfg, ok := d.cfg[id]; ok {
				onewireSensors[pos].DSConfig = append(onewireSensors[pos].DSConfig, cfg.Config())
			} else {
				Log.Debug("id not found before: ", id)
			}
		}
		pos++
	}
	return onewireSensors, nil
}

func (d *DSHandler) Open() {
	if d.sensors == nil {
		return
	}

	d.cfg = make(map[string]*dsSensor.Sensor)
	for _, sensors := range d.sensors {
		for _, sensor := range sensors {
			d.cfg[sensor.ID()] = dsSensor.New(sensor)
		}
	}

}

func (d *DSHandler) Close() {
	for _, s := range d.cfg {
		_ = s.StopPoll()
	}

}
