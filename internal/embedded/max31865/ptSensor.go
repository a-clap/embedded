/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"github.com/a-clap/iot/internal/embedded/avg"
	"github.com/a-clap/iot/internal/embedded/models"
	"sync/atomic"
	"time"
)

type PTHandler interface {
	Poll(data chan Readings, pollTime time.Duration) (err error)
	Temperature() (float32, error)
	Close() error
	ID() string
}

type PTSensor struct {
	handler     PTHandler
	polling     atomic.Bool
	cfg         models.PTConfig
	readings    chan Readings
	temperature models.Temperature
	average     *avg.Avg[float32]
}

var (
	_ models.PTSensor = (*PTSensor)(nil)
)

func NewPTSensor(handler PTHandler) (*PTSensor, error) {
	average, err := avg.New[float32](5)
	if err != nil {
		return nil, err
	}

	return &PTSensor{
		handler: handler,
		polling: atomic.Bool{},
		cfg: models.PTConfig{
			ID:      handler.ID(),
			Enabled: false,
			Samples: 5,
		},
		readings: nil,
		temperature: models.Temperature{
			ID:          handler.ID(),
			Enabled:     false,
			Temperature: 0,
			Stamp:       time.Time{},
		},
		average: average,
	}, nil

}

func (p *PTSensor) Temperature() models.Temperature {
	p.temperature.Temperature = p.average.Average()
	p.temperature.Enabled = p.polling.Load()
	return p.temperature
}

func (p *PTSensor) Poll() error {
	if p.polling.Load() {
		return ErrAlreadyPolling
	}

	p.readings = make(chan Readings, 5)
	if err := p.handler.Poll(p.readings, -1); err != nil {
		return err
	}

	p.cfg.Enabled = true
	p.polling.Store(true)
	go p.handleReadings()

	return nil
}

func (p *PTSensor) StopPoll() error {
	if !p.polling.Load() {
		return ErrNotPolling
	}
	p.cfg.Enabled = false
	defer func() {
		p.polling.Store(false)
		close(p.readings)
	}()
	return p.handler.Close()
}

func (p *PTSensor) Config() models.PTConfig {
	return p.cfg
}

func (p *PTSensor) SetConfig(cfg models.PTConfig) (err error) {
	if err = p.average.Resize(cfg.Samples); err != nil {
		return err
	}
	p.cfg.Samples = cfg.Samples

	if p.cfg.Enabled != cfg.Enabled {
		if cfg.Enabled {
			err = p.Poll()
		} else {
			err = p.StopPoll()
		}
	}
	return
}

func (p *PTSensor) handleReadings() {
	for data := range p.readings {
		p.temperature = models.Temperature{
			ID:      data.ID(),
			Enabled: p.polling.Load(),
			Stamp:   data.Stamp(),
		}
		p.average.Add(data.Temperature())
	}
}
func (p *PTSensor) ID() string {
	return p.handler.ID()
}
