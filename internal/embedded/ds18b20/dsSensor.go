package ds18b20

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/avg"
	"github.com/a-clap/iot/internal/embedded/models"
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
var _ models.DSSensor = (*DSSensor)(nil)

type DSSensor struct {
	handler     DSHandler
	polling     atomic.Bool
	cfg         models.DSConfig
	readings    chan Readings
	temperature models.Temperature
	average     *avg.Avg[float32]
}

func NewDSSensor(handler DSHandler) *DSSensor {
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

	return &DSSensor{
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

func (d *DSSensor) Temperature() models.Temperature {
	d.temperature.Temperature = d.average.Average()
	d.temperature.Enabled = d.polling.Load()
	return d.temperature
}

func (d *DSSensor) Poll() error {
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

func (d *DSSensor) StopPoll() error {
	if !d.polling.Load() {
		return ErrNotPolling
	}
	d.cfg.Enabled = false
	defer func() {
		d.polling.Store(false)
		close(d.readings)
	}()
	return d.handler.Close()
}

func (d *DSSensor) Config() models.DSConfig {
	return d.cfg
}

func (d *DSSensor) SetConfig(cfg models.DSConfig) (err error) {
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
			err = d.Poll()
		} else {
			err = d.StopPoll()
		}
	}
	return
}

func (d *DSSensor) handleReadings() {
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
