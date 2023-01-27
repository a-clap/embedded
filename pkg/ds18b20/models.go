package ds18b20

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/a-clap/iot/pkg/avg"
	"strconv"
	"sync/atomic"
	"time"
)

type DSHandler interface {
	Poll(data chan Readings, pollTime time.Duration) (err error)
	Resolution() (r models.DSResolution, err error)
	SetResolution(res models.DSResolution) error
	Close() error
	ID() string
}

var (
	ErrNotPolling = errors.New("not polling")
)

var _ DSHandler = (*Sensor)(nil)
var _ models.DSSensor = (*ModelsSensor)(nil)

type ModelsSensor struct {
	handler     DSHandler
	polling     atomic.Bool
	cfg         models.DSConfig
	readings    chan Readings
	temperature models.Temperature
	average     *avg.Avg[float32]
}

func NewModels(handler DSHandler) *ModelsSensor {
	id := handler.ID()
	res, err := handler.Resolution()
	if err != nil {
		res = models.Resolution11BIT
	}

	pollTime := func(r models.DSResolution) uint {
		switch r {
		case models.Resolution9BIT:
			return 94
		case models.Resolution10BIT:
			return 188
		case models.Resolution11BIT:
			return 375
		default:
			fallthrough
		case models.Resolution12BIT:
			return 750
		}
	}

	average, err := avg.New[float32](models.DefaultSamples)
	if err != nil {
		panic(err)
	}

	return &ModelsSensor{
		polling: atomic.Bool{},
		handler: handler,
		cfg: models.DSConfig{
			ID:             id,
			Enabled:        false,
			Resolution:     res,
			PollTimeMillis: pollTime(res),
			Samples:        models.DefaultSamples,
		},
		temperature: models.Temperature{
			ID:          id,
			Enabled:     false,
			Temperature: 0,
			Stamp:       time.Time{},
		},
		readings: nil,
		average:  average,
	}
}

func (m *ModelsSensor) Temperature() models.Temperature {
	m.temperature.Temperature = m.average.Average()
	m.temperature.Enabled = m.polling.Load()
	return m.temperature
}

func (d *ModelsSensor) Poll() error {
	if d.polling.Load() {
		return ErrAlreadyPolling
	}

	d.readings = make(chan Readings, 5)
	if err := d.handler.Poll(d.readings, time.Duration(d.cfg.PollTimeMillis)*time.Millisecond); err != nil {
		return err
	}

	d.cfg.Enabled = true
	d.polling.Store(true)
	go d.handleReadings()

	return nil
}

func (d *ModelsSensor) StopPoll() error {
	if !d.polling.Load() {
		return ErrNotPolling
	}
	d.cfg.Enabled = false
	defer d.polling.Store(false)
	return d.handler.Close()
}

func (d *ModelsSensor) Config() models.DSConfig {
	return d.cfg
}

func (d *ModelsSensor) SetConfig(cfg models.DSConfig) (err error) {
	if d.cfg.Resolution != cfg.Resolution {
		if err = d.handler.SetResolution(cfg.Resolution); err != nil {
			return
		}
		d.cfg.Resolution = cfg.Resolution
	}

	d.cfg.PollTimeMillis = cfg.PollTimeMillis
	if err = d.average.Resize(cfg.Samples); err != nil {
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

func (d *ModelsSensor) handleReadings() {
	for data := range d.readings {
		d.temperature = models.Temperature{
			ID:      data.ID(),
			Enabled: d.polling.Load(),
			Stamp:   data.Stamp(),
		}
		if f, err := strconv.ParseFloat(data.Temperature(), 32); err == nil {
			d.average.Add(float32(f))
		}

	}
}
