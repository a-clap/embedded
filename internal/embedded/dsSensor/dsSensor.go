package dsSensor

import (
	"errors"
	. "github.com/a-clap/iot/internal/embedded/logger"
	"github.com/a-clap/iot/pkg/avg"
	"sync/atomic"
	"time"
)

var (
	ErrAlreadyPolling = errors.New("already polling")
	ErrNotPolling     = errors.New("not polling")
)

type PollData interface {
	ID() string
	Temperature() float32
	Stamp() time.Time
	Error() error
}

type Handler interface {
	ID() string

	Resolution() (Resolution, error)
	SetResolution(resolution Resolution) error

	PollTime() uint
	SetPollTime(duration uint) error

	Poll(data chan PollData, timeMillis uint) error
	StopPoll() error
}

type Status struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Temperature float32   `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
}

type Config struct {
	ID             string     `json:"id"`
	Enabled        bool       `json:"enabled"`
	Resolution     Resolution `json:"resolution"`
	PollTimeMillis uint       `json:"poll_time_millis"`
	Samples        uint       `json:"samples"`
}

type Sensor struct {
	polling      atomic.Bool
	handler      Handler
	cfg          Config
	tempReadings chan PollData
	lastRead     Status
	temps        *avg.Avg[float32]
}

type Resolution int

const (
	Resolution9BIT  Resolution = 9
	Resolution10BIT Resolution = 10
	Resolution11BIT Resolution = 11
	Resolution12BIT Resolution = 12
	DefaultSamples  uint       = 5
)

func New(handler Handler) *Sensor {
	id := handler.ID()
	res, err := handler.Resolution()
	if err != nil {
		Log.Debug("resolution read failed on handler: ", id, ", error: ", err)
		res = Resolution11BIT
	}

	pollTime := func(r Resolution) uint {
		switch r {
		case Resolution9BIT:
			return 94
		case Resolution10BIT:
			return 188
		case Resolution11BIT:
			return 375
		default:
			Log.Debug("unspecified resolution: ", r)
			fallthrough
		case Resolution12BIT:
			return 750
		}
	}

	temps, err := avg.New[float32](DefaultSamples)
	if err != nil {
		panic(err)
	}

	return &Sensor{
		polling: atomic.Bool{},
		handler: handler,
		cfg: Config{
			ID:             id,
			Enabled:        false,
			Resolution:     res,
			PollTimeMillis: pollTime(res),
			Samples:        DefaultSamples,
		},
		lastRead: Status{
			ID:          id,
			Enabled:     false,
			Temperature: 0,
			Stamp:       time.Time{},
		},
		tempReadings: nil,
		temps:        temps,
	}
}

func (d *Sensor) Status() Status {
	d.lastRead.Temperature = d.temps.Average()
	d.lastRead.Enabled = d.polling.Load()
	return d.lastRead
}

func (d *Sensor) Poll() error {
	if d.polling.Load() {
		return ErrAlreadyPolling
	}

	d.tempReadings = make(chan PollData, 5)
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

func (d *Sensor) Config() Config {
	return d.cfg
}

func (d *Sensor) SetConfig(cfg Config) (err error) {
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
		d.lastRead = Status{
			ID:      data.ID(),
			Enabled: d.polling.Load(),
			Stamp:   data.Stamp(),
		}
		d.temps.Add(data.Temperature())
	}
}
