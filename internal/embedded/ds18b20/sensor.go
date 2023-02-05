/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

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

var _ DSHandler = (*Handler)(nil)
var _ models.DSSensor = (*Sensor)(nil)

type Sensor struct {
	handler     DSHandler
	polling     atomic.Bool
	cfg         models.DSConfig
	readings    chan Readings
	temperature models.Temperature
	average     *avg.Avg[float32]
}

func NewDSSensor(handler DSHandler) *Sensor {
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

func (s *Sensor) Temperature() models.Temperature {
	s.temperature.Temperature = s.average.Average()
	s.temperature.Enabled = s.polling.Load()
	return s.temperature
}

func (s *Sensor) Poll() error {
	if s.polling.Load() {
		return ErrAlreadyPolling
	}

	s.readings = make(chan Readings, 5)
	if err := s.handler.Poll(s.readings, time.Duration(s.cfg.PollTimeMillis)*time.Millisecond); err != nil {
		return err
	}

	s.cfg.Enabled = true
	s.polling.Store(true)
	go s.handleReadings()

	return nil
}

func (s *Sensor) StopPoll() error {
	if !s.polling.Load() {
		return ErrNotPolling
	}
	s.cfg.Enabled = false
	defer func() {
		s.polling.Store(false)
		close(s.readings)
	}()
	return s.handler.Close()
}

func (s *Sensor) Config() models.DSConfig {
	return s.cfg
}

func (s *Sensor) SetConfig(cfg models.DSConfig) (err error) {
	if s.cfg.Resolution != cfg.Resolution {
		if err = s.handler.SetResolution(cfg.Resolution); err != nil {
			return
		}
		s.cfg.Resolution = cfg.Resolution
	}

	s.cfg.PollTimeMillis = cfg.PollTimeMillis
	if err = s.average.Resize(cfg.Samples); err != nil {
		return err
	}
	s.cfg.Samples = cfg.Samples

	if s.cfg.Enabled != cfg.Enabled {
		if cfg.Enabled {
			err = s.Poll()
		} else {
			err = s.StopPoll()
		}
	}
	return
}

func (s *Sensor) handleReadings() {
	for data := range s.readings {
		if err := data.Error(); err != nil {
			Log.Errorf("readings error, DS id: %s, error is %v\n", data.ID(), err)
		}

		s.temperature = models.Temperature{
			ID:      data.ID(),
			Enabled: s.polling.Load(),
			Stamp:   data.Stamp(),
		}
		f, err := strconv.ParseFloat(data.Temperature(), 32)
		if err != nil {
			Log.Error("parseFloat error on string  %v, error is %v\n", data.Temperature(), err)
		}

		s.average.Add(float32(f))
	}
}
