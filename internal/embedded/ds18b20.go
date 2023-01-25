package embedded

import (
	"errors"
)

type DS18B20Resolution int
type OnewireBusName string

const (
	DS18B20Resolution_9BIT  DS18B20Resolution = 9
	DS18B20Resolution_10BIT DS18B20Resolution = 10
	DS18B20Resolution_11BIT DS18B20Resolution = 11
	DS18B20Resolution_12BIT DS18B20Resolution = 12
)

var (
	ErrNoSuchSensor   = errors.New("specified handler doesnt' exist")
	ErrAlreadyPolling = errors.New("already polling")
	ErrNotPolling     = errors.New("not polling")
)

type BusConfig struct {
	Resolution     DS18B20Resolution `json:"resolution"`
	PollTimeMillis uint              `json:"poll_time_millis"`
	Samples        uint              `json:"samples"`
}

type DSConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	BusConfig
}

type OnewireSensors struct {
	Bus      OnewireBusName `json:"bus"`
	DSConfig []DSConfig     `json:"ds18b20"`
}

type DSHandler struct {
	sensors map[OnewireBusName][]DSSensorHandler
	cfg     map[string]*dsSensor
}

func (d *DSHandler) ConfigSensor(cfg DSConfig) (newConfig DSConfig, err error) {
	ds, err := d.sensorBy(cfg.ID)
	if err != nil {
		return
	}

	if err = ds.setConfig(cfg); err != nil {
		return
	}

	return ds.config(), nil
}

func (d *DSHandler) SensorStatus(id string) (DSConfig, error) {
	s, err := d.sensorBy(id)
	if err != nil {
		return DSConfig{}, err
	}
	return s.config(), nil
}

func (d *DSHandler) sensorBy(id string) (*dsSensor, error) {
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
		onewireSensors[pos].DSConfig = make([]DSConfig, 0, len(v))
		for _, sensor := range v {
			id := sensor.ID()
			if cfg, ok := d.cfg[id]; ok {
				onewireSensors[pos].DSConfig = append(onewireSensors[pos].DSConfig, cfg.config())
			} else {
				log.Debug("id not found before: ", id)
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

	d.cfg = make(map[string]*dsSensor)
	for _, sensors := range d.sensors {
		for _, sensor := range sensors {
			d.cfg[sensor.ID()] = newDsSensor(sensor)
		}
	}

}

func (d *DSHandler) Close() {
	for _, s := range d.cfg {
		_ = s.stopPoll()
	}

}
