package embedded

import (
	"errors"
	"sync/atomic"
	"time"
)

type TemperatureReadings interface {
	ID() string
	Temperature() float32
	Stamp() time.Time
	Error() error
}

type DSSensorHandler interface {
	ID() string

	Resolution() (DS18B20Resolution, error)
	SetResolution(resolution DS18B20Resolution) error

	PollTime() uint
	SetPollTime(duration uint) error

	Poll(data chan TemperatureReadings, t uint) error
	StopPoll() error
}

type DSReadings struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Temperature float32   `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
	Error       error     `json:"error"`
}

type dsSensor struct {
	polling      atomic.Bool
	handler      DSSensorHandler
	cfg          DSConfig
	tempReadings chan TemperatureReadings
	samplesPos   int
	samples      []DSReadings
}

func (d *dsSensor) readings() DSReadings {
	return DSReadings{}
}

func (d *dsSensor) poll() error {
	if d.polling.Load() {
		return ErrAlreadyPolling
	}
	d.tempReadings = make(chan TemperatureReadings, 5)
	if err := d.handler.Poll(d.tempReadings, d.cfg.PollTimeMillis); err != nil {
		return err
	}

	d.cfg.Enabled = true
	d.polling.Store(true)
	go d.handleReadings()

	return nil
}

func (d *dsSensor) stopPoll() error {
	if !d.polling.Load() {
		return ErrNotPolling
	}
	d.cfg.Enabled = false
	defer d.polling.Store(false)
	return d.handler.StopPoll()
}

func (d *dsSensor) config() DSConfig {
	return d.cfg
}

func (d *dsSensor) setConfig(cfg DSConfig) (err error) {
	if d.cfg.Resolution != cfg.Resolution {
		if err = d.handler.SetResolution(cfg.Resolution); err != nil {
			return
		}
		d.cfg.Resolution = cfg.Resolution
	}

	if d.cfg.PollTimeMillis != cfg.PollTimeMillis {
		if err = d.handler.SetPollTime(cfg.PollTimeMillis); err != nil {
			return
		}
	}

	if cfg.Samples == 0 {
		return errors.New("wrong samples value")
	}

	if d.cfg.Enabled != cfg.Enabled {
		if cfg.Enabled {
			if err = d.poll(); err != nil {
				return
			}
		} else {
			if err = d.stopPoll(); err != nil {
				return
			}
		}
	}
	return
}

func (d *dsSensor) handleReadings() {
	//for data := range d.tempReadings {
	//	if cap(d.samples) != int(d.cfg.Samples) {
	//
	//	}
	//
	//	r := DSReadings{
	//		ID:          data.ID(),
	//		Enabled:     d.polling.Load(),
	//		Temperature: data.Temperature(),
	//		Stamp:       data.Stamp(),
	//		Error:       data.Error(),
	//	}
	//
	//}
}
