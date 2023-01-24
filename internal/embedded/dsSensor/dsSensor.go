package dsSensor

import (
	"errors"
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/a-clap/iot/pkg/avg"
	"sync/atomic"
	"time"
)

var (
	ErrAlreadyPolling = errors.New("already polling")
	ErrNotPolling     = errors.New("not polling")
)

var _ models.DSSensor = (*Sensor)(nil)

type Sensor struct {
	polling      atomic.Bool
	handler      models.Handler
	cfg          models.DSConfig
	tempReadings chan models.PollData
	lastRead     models.DSStatus
	temps        *avg.Avg[float32]
}

func New(handler models.Handler) *Sensor {
	id := handler.ID()
	res, err := handler.Resolution()
	if err != nil {
		Log.Debug("resolution read failed on handler: ", id, ", error: ", err)
		res = models.Resolution11BIT
	}

	pollTime := func(r models.Resolution) uint {
		switch r {
		case models.Resolution9BIT:
			return 94
		case models.Resolution10BIT:
			return 188
		case models.Resolution11BIT:
			return 375
		default:
			Log.Debug("unspecified resolution: ", r)
			fallthrough
		case models.Resolution12BIT:
			return 750
		}
	}

	temps, err := avg.New[float32](models.DefaultSamples)
	if err != nil {
		panic(err)
	}

	return &Sensor{
		polling: atomic.Bool{},
		handler: handler,
		cfg: models.DSConfig{
			ID:             id,
			Enabled:        false,
			Resolution:     res,
			PollTimeMillis: pollTime(res),
			Samples:        models.DefaultSamples,
		},
		lastRead: models.DSStatus{
			ID:          id,
			Enabled:     false,
			Temperature: 0,
			Stamp:       time.Time{},
		},
		tempReadings: nil,
		temps:        temps,
	}
}

func (d *Sensor) Status() models.DSStatus {
	d.lastRead.Temperature = d.temps.Average()
	d.lastRead.Enabled = d.polling.Load()
	return d.lastRead
}

func (d *Sensor) Poll() error {
	if d.polling.Load() {
		return ErrAlreadyPolling
	}

	d.tempReadings = make(chan models.PollData, 5)
	if err := d.handler.Poll(d.tempReadings, d.cfg.PollTimeMillis); err != nil {
		return err
	}

	d.cfg.Enabled = true
	d.polling.Store(true)
	go d.handleReadings()

	return nil
}

func (d *Sensor) StopPoll() error {
	if !d.polling.Load() {
		return ErrNotPolling
	}
	d.cfg.Enabled = false
	defer d.polling.Store(false)
	return d.handler.StopPoll()
}

func (d *Sensor) Config() models.DSConfig {
	return d.cfg
}

func (d *Sensor) SetConfig(cfg models.DSConfig) (err error) {
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
		d.cfg.PollTimeMillis = cfg.PollTimeMillis
	}

	if err = d.temps.Resize(cfg.Samples); err != nil {
		return err
	}
	d.cfg.Samples = cfg.Samples

	if d.cfg.Enabled != cfg.Enabled {
		if cfg.Enabled {
			if err = d.Poll(); err != nil {
				return
			}
		} else {
			if err = d.StopPoll(); err != nil {
				return
			}
		}
	}
	return
}

func (d *Sensor) handleReadings() {
	for data := range d.tempReadings {
		d.lastRead = models.DSStatus{
			ID:      data.ID(),
			Enabled: d.polling.Load(),
			Stamp:   data.Stamp(),
		}
		d.temps.Add(data.Temperature())
	}
}
