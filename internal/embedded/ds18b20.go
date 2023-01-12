package embedded

import (
	"errors"
	"time"
)

const (
	DS18B20Resolution_9BIT  = 9
	DS18B20Resolution_10BIT = 10
	DS18B20Resolution_11BIT = 11
	DS18B20Resolution_12BIT = 12
)

var (
	ErrInvalidBus = errors.New("specified bus doesn't exist")
)

type DS18B20Resolution int
type OnewireBusName string

type TemperatureReadings interface {
	ID() string
	Temperature() float32
	Stamp() time.Time
	Error() error
}

type DS18B20Sensor interface {
	ID() string
	Resolution() (DS18B20Resolution, error)
	SetResolution(resolution DS18B20Resolution) error
	Poll(data chan TemperatureReadings, pollTimeMillis uint64) error
	StopPoll() error
}

type DSConfig struct {
	sensor         DS18B20Sensor
	ID             string            `json:"id"`
	Enabled        bool              `json:"enabled"`
	Resolution     DS18B20Resolution `json:"resolution"`
	PollTimeMillis uint              `json:"poll_time_millis"`
	Samples        uint              `json:"samples"`
}

type OnewireSensors struct {
	Bus      OnewireBusName `json:"bus"`
	DSConfig []DSConfig     `json:"ds18b20"`
}

type DSHandler struct {
	sensors map[OnewireBusName][]DS18B20Sensor
	cfg     map[string]DSConfig
}

func (d *DSHandler) defaultPollTime(r DS18B20Resolution) uint {
	switch r {
	case DS18B20Resolution_9BIT:
		return 94
	case DS18B20Resolution_10BIT:
		return 188
	case DS18B20Resolution_11BIT:
		return 375
	default:
		log.Debug("unspecified resolution: ", r)
		fallthrough
	case DS18B20Resolution_12BIT:
		return 750
	}
}

func (d *DSHandler) init() {
	if d.sensors == nil {
		return
	}

	d.cfg = make(map[string]DSConfig)
	for _, sensors := range d.sensors {
		for _, sensor := range sensors {
			id := sensor.ID()
			res, err := sensor.Resolution()
			if err != nil {
				log.Debug("resolution read failed on sensor: ", id, ", error: ", err)
				res = DS18B20Resolution_11BIT
			}
			d.cfg[id] = DSConfig{
				ID:             id,
				Enabled:        false,
				Resolution:     res,
				PollTimeMillis: d.defaultPollTime(res),
				Samples:        1,
			}
		}
	}
}

//func (h *Handler) SetResolutionForBus(name OnewireBusName, resolution DS18B20Resolution) error {
//	sensors, ok := h.ds.sensors[name]
//	if !ok {
//		return ErrInvalidBus
//	}
//	for _, sensor := range sensors {
//		id := sensor.ID()
//		ds, ok := h.ds.cfg[id]
//		if !ok {
//			log.Debug("unknown sensor ", id)
//		}
//	}
//
//}

func (d *DSHandler) Status() ([]OnewireSensors, error) {
	onewireSensors := make([]OnewireSensors, len(d.sensors))

	pos := 0
	for k, v := range d.sensors {
		onewireSensors[pos].Bus = k
		onewireSensors[pos].DSConfig = make([]DSConfig, 0, len(v))
		for _, sensor := range v {
			id := sensor.ID()
			if cfg, ok := d.cfg[id]; ok {
				onewireSensors[pos].DSConfig = append(onewireSensors[pos].DSConfig, cfg)
			} else {
				log.Debug("id not found before: ", id)
			}
		}
		pos++
	}
	return onewireSensors, nil
}
